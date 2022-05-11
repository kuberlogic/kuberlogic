/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package cli

import "github.com/spf13/cobra"

func getString(cmd *cobra.Command, flag string) (value *string, err error) {
	if cmd.Flags().Changed(flag) {
		value, err := cmd.Flags().GetString(flag)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	return
}

func setInt64(cmd *cobra.Command, flag string) (value *int64, err error) {
	if cmd.Flags().Changed(flag) {
		value, err := cmd.Flags().GetInt64(flag)
		if err != nil {
			return nil, err
		}
		return &value, nil
	}
	return
}
