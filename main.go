package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/Khan/genqlient/graphql"
	"github.com/getsynq/connections-tableau/internal"
	"github.com/getsynq/connections-tableau/metadata"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"net/url"
	"os"
	"strings"
	"time"
)

var TableauUrl string
var TableauSite string
var TableauTokenName string
var TableauTokenValue string

var rootCmd = &cobra.Command{
	Use:   "connections-tableau",
	Short: "Small utility to collect Tableau information which is only available with Admin permissions",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&TableauUrl, "url", "", "Full URL of Tableau (e.g. `https://prod-uk-a.online.tableau.com`)")
	rootCmd.PersistentFlags().StringVar(&TableauSite, "site", "", "Site name (e.g. `synqtest` from https://prod-uk-a.online.tableau.com/t/synqtest/)")
	rootCmd.PersistentFlags().StringVar(&TableauTokenName, "token_name", "", "Name of the Private Access Token (e.g. `synq`)")
	rootCmd.PersistentFlags().StringVar(&TableauTokenValue, "token", "", "Value of Personal Access Token for Tableau with Admin permissions")

	rootCmd.PreRunE = func(cmd *cobra.Command, args []string) error {

		if TableauUrl == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Full URL of Tableau (e.g. `https://prod-uk-a.online.tableau.com`)",
			}, &TableauUrl, survey.WithValidator(internal.UrlValidator))
			if err != nil {
				return err
			}
		}

		if TableauSite == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Site name (e.g. `synqtest` from https://prod-uk-a.online.tableau.com/t/synqtest/), leave empty for self-hosted Tableau",
			}, &TableauSite)
			if err != nil {
				return err
			}
		}

		if TableauTokenName == "" {
			err := survey.AskOne(&survey.Input{
				Message: "Name of the Private Access Token",
				Default: "synq",
			}, &TableauTokenName, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		}

		if TableauTokenValue == "" {
			err := survey.AskOne(&survey.Password{
				Message: "Value of Personal Access Token for Tableau with Admin permissions",
			}, &TableauTokenValue, survey.WithValidator(survey.Required))
			if err != nil {
				return err
			}
		}

		if TableauUrl == "" || TableauTokenName == "" || TableauTokenValue == "" {
			cmd.Help()
			return errors.New("Not all required parameters provided")
		}
		return nil
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {

		TableauUrl = cleanupUrl(TableauUrl)

		TableauApiVersion, err := internal.GetVersion(TableauUrl)
		if err != nil {
			return errors.Wrap(err, "failed to obtain Tableau API version")
		}

		token, _, err := internal.LoginPersonalAccessToken(TableauUrl, TableauApiVersion, TableauSite, TableauTokenName, TableauTokenValue)

		if err != nil {
			panic(errors.Wrap(err, "failed to authenticate"))
		}

		ctx := context.Background()

		client := graphql.NewClient(fmt.Sprintf("%s/api/metadata/graphql", TableauUrl), internal.HttpClientWithToken(token))

		acceptConnectionTypes := map[string]bool{"bigquery": true, "snowflake": true, "redshift": true, "clickhouse": true}
		databaseTables := make([]*metadata.GetDatabaseTablesDefinitionsDatabaseTablesConnectionNodesDatabaseTable, 0)

		perPage := 100
		page := 0
		totalPages := 0

		for {
			resp, err := metadata.GetDatabaseTablesDefinitions(ctx, client, perPage, perPage*page)
			if err != nil {
				panic(errors.Wrap(err, "failed to obtain metadata"))
			}
			if totalPages == 0 {
				totalPages = resp.DatabaseTablesConnection.TotalCount / perPage
			}
			for _, databaseTable := range resp.DatabaseTablesConnection.Nodes {
				databaseTable := databaseTable
				if acceptConnectionTypes[databaseTable.ConnectionType] {
					databaseTables = append(databaseTables, &databaseTable)
				}
			}
			page += 1
			if page > totalPages {
				break
			}
		}

		fmt.Printf("Discovered %d database tables\n", len(databaseTables))

		jsonBytes, err := json.MarshalIndent(databaseTables, "", "  ")
		if err != nil {
			panic(errors.Wrap(err, "failed to create json"))
		}

		fileName := strings.ReplaceAll(fmt.Sprintf("tables-%s.json", time.Now().UTC().Format(time.RFC3339)), ":", "_")
		err = os.WriteFile(fileName, jsonBytes, 0644)
		if err != nil {
			return errors.Wrapf(err, "failed to write file %s", fileName)
		}

		fmt.Printf("File %s created\n", fileName)

		return nil
	}

}

func cleanupUrl(in string) string {
	u, err := url.Parse(in)
	if err != nil {
		panic(fmt.Sprintf("failed to parse url %s", in))
	}
	u.Fragment = ""
	return u.String()
}

//go:generate go run github.com/Khan/genqlient
func main() {

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}
