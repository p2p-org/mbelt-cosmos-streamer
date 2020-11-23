package commands

import (
	"log"

	"github.com/p2p-org/mbelt-cosmos-streamer/cmd/worker"
	"github.com/p2p-org/mbelt-cosmos-streamer/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/jinzhu/configor"
)

var (
	// Used for flags.
	sync             bool
	syncForce        bool
	updHead          bool
	syncFrom         int
	syncFromDbOffset int

	conf config.Config

	rootCmd = &cobra.Command{
		Use:   "mbelt-cosmos-streamer",
		Short: "A streamer of cosmos's entities to PostgreSQL DB through Kafka",
		Long: `This app synchronizes with current cosmos state and keeps in sync by subscribing on it's updates.
Entities (tipsets, blocks and messages) are being pushed to Kafka. There are also sinks that get
those entities from Kafka streams and push them in PostgreSQL DB.'`,
		Run: func(cmd *cobra.Command, args []string) {
			worker.Start(&conf, sync, syncForce, updHead, syncFrom, syncFromDbOffset)
		},
	}
)

func init() {
	if err := configor.Load(&conf); err != nil {
		log.Fatal(err)
	}

	rootCmd.PersistentFlags().BoolVarP(&sync, "sync", "s", true,
		"Turn on sync starting from last block in DB")
	rootCmd.PersistentFlags().BoolVarP(&syncForce, "sync-force", "f", false,
		"Turn on sync starting from genesis block")
	rootCmd.PersistentFlags().BoolVarP(&updHead, "sub-head-updates", "u", true,
		"Turn on subscription on head updates")
	rootCmd.PersistentFlags().IntVarP(&syncFrom, "sync-from", "F", -1,
		"Height to start sync from. Dont provide or provide negative number to sync from max height in DB")
	rootCmd.PersistentFlags().IntVarP(&syncFromDbOffset, "sync-from-db-offset", "o", 100,
		"Specify offset from max height in DB to start sync from (maxHeightInDb - offset)")

	viper.BindPFlag("sync", rootCmd.PersistentFlags().Lookup("sync"))
	viper.BindPFlag("sync_force", rootCmd.PersistentFlags().Lookup("sync-force"))
	viper.BindPFlag("sub_head_updates", rootCmd.PersistentFlags().Lookup("sub-head-updates"))
	viper.BindPFlag("sync_from", rootCmd.PersistentFlags().Lookup("sync-from"))
	viper.BindPFlag("sync_from_db_offset", rootCmd.PersistentFlags().Lookup("sync-from-db-offset"))
	viper.SetDefault("sync", true)
	viper.SetDefault("sync_force", false)
	viper.SetDefault("sub_head_updates", true)
	viper.SetDefault("sync_from", -1)
	viper.SetDefault("sync_from_db_offset", 100)
}

func Execute() error {
	return rootCmd.Execute()
}
