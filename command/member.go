package command

/*
MIT License

Copyright (c) 2025 Juan Carlos Daille

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
import (
	"NodoCb/global"
	"NodoCb/manager/configuration"
	"NodoCb/manager/database"
	"NodoCb/util"
	"fmt"

	"github.com/spf13/cobra"
)

var memberCmd = &cobra.Command{
	Use:   "member",
	Short: "Administra los miembros",
	Long:  `.`,
	Args:  cobra.RangeArgs(0, 1),
	Run: func(cmd *cobra.Command, args []string) {
		if parameter_new {
			CreateMember(parameter_organization)
		} else if parameter_list {
			ListMember()
		}
	},
}

func init() {
	memberCmd.Flags().BoolVar(&parameter_new, "new", false, "create a new member")
	memberCmd.Flags().BoolVar(&parameter_list, "list", false, "list member")
	memberCmd.Flags().StringVar(&parameter_organization, "organization", "", "The name of the organization")
	/*memberCmd.Flags().StringVar(&parameter_name, "name", "", "The name of the network (required)")
	memberCmd.Flags().StringVar(&parameter_description, "description", "", "The description of the network (required)")
	memberCmd.Flags().StringVar(&parameter_port, "port", "", "The port where others will joins to this node (required)")
	memberCmd.Flags().StringVar(&parameter_networkkey, "networkkey", "", "The key used to connect to the network")
	memberCmd.Flags().StringVar(&parameter_endpoint, "endpoint", "", "The firt endpoint to join to the network")
	memberCmd.Flags().StringVar(&parameter_country, "country", "", "country of the network")*/
	/*createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("description")
	createCmd.MarkFlagRequired("port")
	createCmd.MarkFlagRequired("endpoint")*/
	rootCmd.AddCommand(memberCmd)
}

func CreateMember(name string) {
	if zip, ok := util.CreateZip(name + ".zip"); ok {
		ca := configuration.LoadX509Certificate("./ca.pem")
		caPriv := configuration.LoadX509PrivateKey("./caPriv.pem")
		certPub, certPriv := createCertificate(ca, caPriv, name)
		util.WriteZipStringFile(zip, "cert.pem", certPub)
		util.WriteZipStringFile(zip, "key.pem", certPriv)
		util.WriteZipStringFile(zip, "config.json", global.BASEConfig)
		util.WriteZipFile(zip, "NodoCb")
		zip.Close()

		// @TODO faltan datos
		database.UpdateMember(global.MemberType{
			Organization: parameter_organization,
			PublicCert:   certPub,
		})
	}
}

func ListMember() {
	for i, j := range database.GetAllMembers() {
		fmt.Println(i, " - ", j)
	}
}
