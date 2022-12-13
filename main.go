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
			survey.AskOne(&survey.Input{
				Message: "Full URL of Tableau (e.g. `https://prod-uk-a.online.tableau.com`)",
			}, &TableauUrl, survey.WithValidator(internal.UrlValidator))
		}

		if TableauSite == "" {
			survey.AskOne(&survey.Input{
				Message: "Site name (e.g. `synqtest` from https://prod-uk-a.online.tableau.com/t/synqtest/)",
			}, &TableauSite, survey.WithValidator(survey.Required))
		}

		if TableauTokenName == "" {
			survey.AskOne(&survey.Input{
				Message: "Name of the Private Access Token",
				Default: "synq",
			}, &TableauTokenName, survey.WithValidator(survey.Required))
		}

		if TableauTokenValue == "" {
			survey.AskOne(&survey.Password{
				Message: "Value of Personal Access Token for Tableau with Admin permissions",
			}, &TableauTokenValue, survey.WithValidator(survey.Required))
		}

		if TableauUrl == "" || TableauSite == "" || TableauTokenName == "" || TableauTokenValue == "" {
			cmd.Help()
			return errors.New("Not all required parameters provided")
		}
		return nil
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {

		token, _, err := internal.LoginPersonalAccessToken(TableauUrl, TableauSite, TableauTokenName, TableauTokenValue)

		if err != nil {
			panic(errors.Wrap(err, "failed to authenticate"))
		}

		ctx := context.Background()

		client := graphql.NewClient(fmt.Sprintf("%s/api/metadata/graphql", TableauUrl), internal.HttpClientWithToken(token))

		acceptConnectionTypes := map[string]bool{"bigquery": true, "snowflake": true, "redshift": true}
		var databaseTables []*metadata.GetDatabaseTablesDefinitionsDatabaseTablesConnectionNodesDatabaseTable

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
				if acceptConnectionTypes[databaseTable.ConnectionType] {
					databaseTables = append(databaseTables, &databaseTable)
				}
			}
			page += 1
			if page > totalPages {
				break
			}
		}

		jsonBytes, err := json.MarshalIndent(databaseTables, "", "  ")
		if err != nil {
			panic(errors.Wrap(err, "failed to create json"))
		}

		fmt.Println("Discovered Database tables:")
		fmt.Println(string(jsonBytes))

		return nil
	}

}

//go:generate go run github.com/Khan/genqlient
func main() {

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}

}
