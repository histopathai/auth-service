// migrate-roles rewrites the role names stored before the user groups existed:
// "user" becomes "pathologist", "viewer" becomes "datascientist".
//
// The services already read the old names correctly (model.UserRole.Normalize),
// so this is a cleanup and can run at any time after they are deployed. Run it
// last: a rollback to an auth-service build that predates the groups does not
// know the new names.
//
//	go run ./cmd/migrate-roles -project histopathai-478716            # dry run: prints, writes nothing
//	go run ./cmd/migrate-roles -project histopathai-478716 -apply     # writes
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/histopathai/auth-service/internal/domain/model"
	"google.golang.org/api/iterator"
)

func main() {
	project := flag.String("project", os.Getenv("PROJECT_ID"), "GCP project id (default: $PROJECT_ID)")
	database := flag.String("database", firestore.DefaultDatabaseID, "Firestore database; auth-service uses the default one")
	collection := flag.String("collection", "users", "users collection")
	apply := flag.Bool("apply", false, "write the changes; without it nothing is written")
	flag.Parse()

	if *project == "" {
		log.Fatal("no project: pass -project or set PROJECT_ID")
	}

	ctx := context.Background()
	client, err := firestore.NewClientWithDatabase(ctx, *project, *database)
	if err != nil {
		log.Fatalf("firestore client: %v", err)
	}
	defer client.Close()

	mode := "DRY RUN — nothing is written"
	if *apply {
		mode = "APPLY — writing"
	}
	fmt.Printf("%s\nproject=%s database=%s collection=%s\n\n", mode, *project, *database, *collection)

	counts := map[model.UserRole]int{}
	changed, failed := 0, 0

	it := client.Collection(*collection).Documents(ctx)
	defer it.Stop()
	for {
		doc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("reading %s: %v", *collection, err)
		}

		stored, _ := doc.Data()["role"].(string)
		old := model.UserRole(stored)
		counts[old]++
		target := old.Normalize()
		if target == old {
			continue
		}

		email, _ := doc.Data()["email"].(string)
		fmt.Printf("  %-28s %-36s %s -> %s\n", doc.Ref.ID, email, old, target)
		changed++
		if !*apply {
			continue
		}

		// The precondition refuses the write if the document changed since it was
		// read — an admin may be editing the same user in the panel right now.
		_, err = doc.Ref.Update(ctx, []firestore.Update{
			{Path: "role", Value: string(target)},
			{Path: "updated_at", Value: time.Now()},
		}, firestore.LastUpdateTime(doc.UpdateTime))
		if err != nil {
			fmt.Printf("    FAILED: %v\n", err)
			failed++
		}
	}

	fmt.Println("\nroles found:")
	for role, n := range counts {
		name := string(role)
		if name == "" {
			name = "(none)"
		}
		fmt.Printf("  %-16s %d\n", name, n)
	}

	switch {
	case !*apply:
		fmt.Printf("\n%d document(s) would change. Re-run with -apply to write.\n", changed)
	case failed > 0:
		fmt.Printf("\n%d of %d write(s) failed — re-run; it only touches what still needs changing.\n", failed, changed)
		os.Exit(1)
	default:
		fmt.Printf("\n%d document(s) changed.\n", changed)
	}
}
