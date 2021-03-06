// Copyright © 2018 EOS Canada <info@eoscanada.com>

package cmd

import (
	"fmt"
	"os"

	eos "github.com/eoscanada/eos-go"
	"github.com/eoscanada/eosc/analysis"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var msigReviewCmd = &cobra.Command{
	Use:   "review [proposer] [proposal name]",
	Short: "Review a proposal in the eosio.msig contract",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		api := getAPI()

		proposer := toAccount(args[0], "proposer")
		proposalName := toName(args[1], "proposal name")

		response, err := api.GetTableRows(
			eos.GetTableRowsRequest{
				Code:       "eosio.msig",
				Scope:      string(proposer),
				Table:      "proposal",
				JSON:       true,
				LowerBound: string(proposalName),
				Limit:      1,
			},
		)
		errorCheck("get table row", err)

		var transactions []struct {
			ProposalName eos.Name     `json:"proposal_name"`
			Transaction  eos.HexBytes `json:"packed_transaction"`
		}
		err = response.JSONToStructs(&transactions)
		errorCheck("reading proposed transactions", err)

		var tx *eos.Transaction
		for _, txData := range transactions {
			if txData.ProposalName == proposalName {
				err := eos.UnmarshalBinary(txData.Transaction, &tx)
				errorCheck("unmarshalling packed transaction", err)

				ana := analysis.NewAnalyzer(viper.GetBool("msig-review-cmd-dump"))
				err = ana.AnalyzeTransaction(tx)
				errorCheck("analyzing", err)

				fmt.Println("Proposer:", proposer)
				fmt.Println("Proposal name:", proposalName)
				fmt.Println()
				os.Stdout.Write(ana.Writer.Bytes())
			}
		}
		if tx == nil {
			errorCheck("transaction", fmt.Errorf("not found"))
		}
	},
}

func init() {
	msigCmd.AddCommand(msigReviewCmd)

	msigReviewCmd.Flags().BoolP("dump", "", true, "Do verbose analysis, and dump more contents of transactions and actions.")

	for _, flag := range []string{"dump"} {
		if err := viper.BindPFlag("msig-review-cmd-"+flag, msigReviewCmd.Flags().Lookup(flag)); err != nil {
			panic(err)
		}
	}

}
