package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/td0m/doorman"
)

const (
	totalUsers          = 100_000
	totalItems          = 100
	totalUserItemTuples = totalUsers * 10
)

func cleanup(ctx context.Context, conn *pgxpool.Pool) {
	query := `
		delete from tuples;
	`
	if _, err := conn.Exec(ctx, query); err != nil {
		panic(err)
	}
}

var rnd = rand.New(rand.NewSource(42))

func randomUser() string {
	i := rand.Intn(totalUsers)
	return "user:" + strconv.Itoa(i)
}

func randomItem() string {
	i := rand.Intn(totalItems)
	return "item:" + strconv.Itoa(i)
}

func makeGroup(s *doorman.Server, obj string, sizes []int) {
	if len(sizes) == 0 {
		return
	}

	size, sizes2 := sizes[0], sizes[1:]

	req := doorman.WriteRequest{
	 	Add: make([]doorman.Tuple, size),
	}
	for i := 0; i < size; i++ {
		tuple := doorman.Tuple{
			Object:   obj,
			Relation: "child",
			Subject:  fmt.Sprintf("group:%d_%d", len(sizes), i),
		}

		if len(sizes) == 1 {
			tuple.Relation = "member"
			tuple.Subject = randomUser()
		}
		req.Add[i] = tuple

		makeGroup(s, tuple.Subject, sizes2)
	}

	_, err := s.Write(context.Background(), req)
	if err != nil {
		panic(err)
	}

}

func run() error {
	ctx := context.Background()
	conn, err := pgxpool.New(ctx, "")
	if err != nil {
		return fmt.Errorf("pg failed: %w", err)
	}

	var write bool
	flag.BoolVar(&write, "write", false, "write")
	flag.Parse()

	schema := doorman.SchemaDef{
		Types: []doorman.SchemaTypeDef{
			{
				Name: "group",
				Relations: []doorman.SchemaRelationDef{
					{Name: "child"},
					{Name: "member", Value: doorman.NewComputedVia("child", "member")},
				},
			},
			{
				Name: "item",
				Relations: []doorman.SchemaRelationDef{
					{Name: "owner"},
					{Name: "viewer"},
					{Name: "can_retrieve", Value: doorman.NewUnion(doorman.NewComputed("owner"), doorman.NewComputed("viewer"))},
					{Name: "foo", Value: doorman.NewUnion(doorman.NewComputed("can_retrieve"))},
					{Name: "bar", Value: doorman.NewUnion(doorman.NewComputed("foo"))},
				},
			},
		},
	}

	srv := doorman.NewServer(schema, doorman.NewTupleStore(conn))

	if write {
		cleanup(ctx, conn)
		start := time.Now()
		//
		// for i := 0; i < totalUserItemTuples; i++ {
		// 	fmt.Println(float64(math.Round(float64(i) / totalUserItemTuples * 100)), "%")
		// 	_, err := srv.Write(ctx, doorman.WriteRequest{
		// 		Add: []doorman.Tuple{
		// 			{Object: randomItem(), Relation: "owner", Subject: randomUser()},
		// 		},
		// 	})
		// 	if err != nil {
		// 		return fmt.Errorf("write failed: %w", err)
		// 	}
		// }
		//

		makeGroup(srv, "group:all", []int{3, 3, 3, 10, 300})
		_, err := srv.Write(ctx, doorman.WriteRequest{
			Add: []doorman.Tuple{
				{Object: randomItem(), Relation: "owner", Subject: "group:all"},
			},
		})
		if err != nil {
			return fmt.Errorf("write failed: %w", err)
		}

		duration := time.Since(start)

		var total int64
		err = conn.QueryRow(ctx, "select count(*) from tuples").Scan(&total)
		if err != nil {
			return err
		}

		fmt.Printf("done writing, took: %+v, %d written. That is %+v / tuple\n", duration, total, duration/time.Duration(total))
	}

	{
		start := time.Now()
		sampleSize := 10

		yesCount := 0

		for i := 0; i < sampleSize; i++ {
			err = srv.Check(ctx, doorman.Tuple{Object: "group:all", Relation: "member", Subject: randomUser()})
			if err == nil {
				yesCount++
			}
		}

		duration := time.Since(start)
		fmt.Printf("done checking, took: %+v. That is %+v / check. %d successful\n", duration, duration/time.Duration(sampleSize), yesCount)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
