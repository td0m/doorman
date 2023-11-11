package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	pb "github.com/td0m/doorman/gen/go"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

var usage = `doorman {{version}}
  The official command-line interface for doorman.

usage:
  doorman command [options]

commands:
	grant          grants subject access to an object via a role.
	revoke         revokes subject access to an object via a role.
	check          checks if the subject can access the object via specified verb.
	roles upsert   creates or updates a role.
`

var (
	srv pb.DoormanClient
)

func app(ctx context.Context) error {
	addr := "localhost:13335"
	if envAddr := os.Getenv("DOORMAN_HOST"); len(envAddr) > 0 {
		addr = envAddr
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("grpc.Dial failed: %w", err)
	}
	defer conn.Close()

	srv = pb.NewDoormanClient(conn)

	cmd := os.Args[1]
	switch cmd {
	case "check":
		if len(os.Args) != 5 {
			return errors.New("usage: check [subject] [verb] [object]")
		}
		subject, verb, object := os.Args[2], os.Args[3], os.Args[4]

		res, err := srv.Check(ctx, &pb.CheckRequest{
			Subject: subject,
			Verb:    verb,
			Object:  object,
		})
		if err != nil {
			return err
		}

		if res.Success {
			fmt.Println("✅")
		} else {
			fmt.Println("X")
		}

	case "revoke":
		if len(os.Args) != 5 {
			return errors.New("usage: revoke [subject] [role] [object]")
		}
		subject, role, object := os.Args[2], os.Args[3], os.Args[4]

		res, err := srv.Revoke(ctx, &pb.RevokeRequest{
			Subject: subject,
			Role:    role,
			Object:  object,
		})
		if err != nil {
			return err
		}
		fmt.Println(res)

	case "grant":
		if len(os.Args) != 5 {
			return errors.New("usage: grant [subject] [role] [object]")
		}
		subject, role, object := os.Args[2], os.Args[3], os.Args[4]

		res, err := srv.Grant(ctx, &pb.GrantRequest{
			Subject: subject,
			Role:    role,
			Object:  object,
		})
		if err != nil {
			return err
		}
		fmt.Println(res)

	case "list-objects":
		if len(os.Args) != 3 {
			return errors.New("usage: list-objects [subject]")
		}
		sub := os.Args[2]

		res, err := srv.ListObjects(ctx, &pb.ListObjectsRequest{
			Subject: sub,
		})
		if err != nil {
			return err
		}
		printRelations(res.Items)

	case "rebuild-cache":
		_, err := srv.RebuildCache(ctx, &pb.RebuildCacheRequest{})
		if err != nil {
			return err
		}

	case "roles":
		os.Args = os.Args[1:]
		switch os.Args[1] {
		case "list":
			if len(os.Args) != 2 {
				return errors.New("usage: roles list")
			}

			res, err := srv.ListRoles(ctx, &pb.ListRolesRequest{})
			if err != nil {
				return err
			}

			printRoles(res.Items)
		case "upsert":
			if len(os.Args) < 3 {
				return errors.New("usage: roles create [id] [verb1] ... [verbN]")
			}
			id, verbs := os.Args[2], os.Args[3:]

			fmt.Println("role", id, verbs)

			role, err := srv.UpsertRole(ctx, &pb.UpsertRoleRequest{
				Id:    id,
				Verbs: verbs,
			})
			if err != nil {
				return fmt.Errorf("upsert failed: %w", err)
			}
			fmt.Println(role)
		default:
			return fmt.Errorf("invalid command: %s", os.Args[1])
		}
	}

	return nil
}

func emojify(id string) string {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) == 1 {
		return id
	}
	typ := parts[0]
	switch typ {
	case "user":
		return "👤" + id
	case "group":
		return "🏘️" + id
	case "post":
		return "🗒️" + id
	default:
		return id
	}
}

func printAttrs(attrs map[string]any) {
	for k, v := range attrs {
		fmt.Printf("%s\t\t%+v\n", k, v)
	}
}

func printRelations(rs []*pb.Relation) {
	rows := [][]string{}
	for _, r := range rs {
		rows = append(rows, []string{emojify(r.Subject), r.Verb, emojify(r.Object)})
	}
	table := table.New().
		Border(lipgloss.NormalBorder()).
		Headers("Subject", "Verb", "Object").
		StyleFunc(func(row, _ int) lipgloss.Style {
			switch row {
			case 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Padding(0, 1)
			default:
				return lipgloss.NewStyle().Padding(0, 1)
			}
		}).
		Rows(rows...)

	fmt.Println(table.Render())
}

func printRoles(rs []*pb.Role) {
	rows := [][]string{}
	for _, r := range rs {
		rows = append(rows, []string{emojify(r.Id), strings.Join(r.Verbs, ", ")})
	}
	table := table.New().
		Border(lipgloss.NormalBorder()).
		Headers("Role", "Verbs").
		StyleFunc(func(row, _ int) lipgloss.Style {
			switch row {
			case 0:
				return lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true).Padding(0, 1)
			default:
				return lipgloss.NewStyle().Padding(0, 1)
			}
		}).
		Rows(rows...)

	fmt.Println(table.Render())
}

func main() {
	usage = strings.Replace(usage, "{{version}}", "v0", 1)
	if len(os.Args) < 2 {
		fmt.Println(usage)
		os.Exit(0)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*8)
	defer cancel()

	if err := app(ctx); err != nil {
		if st, ok := status.FromError(err); ok {
			fmt.Printf("(%s) %s\n", st.Code(), st.Message())
		} else {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}
