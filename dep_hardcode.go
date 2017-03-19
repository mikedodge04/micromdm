package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/micromdm/mdm"
	"github.com/micromdm/nano/checkin"
	"github.com/micromdm/nano/command"
	"github.com/pkg/errors"
)

func hardcodeCommands(sm *config) error {
	sub := sm.pubclient
	cmdsvc := sm.commandService
	pushsvc := sm.pushService
	authEvents, err := sub.Subscribe("hardcode-dep", checkin.AuthenticateTopic)
	if err != nil {
		return errors.Wrapf(err,
			"subscribing devices to %s topic", checkin.AuthenticateTopic)
	}

	go func() {
		for {
			select {
			case event := <-authEvents:
				var ev checkin.Event
				if err := checkin.UnmarshalEvent(event.Message, &ev); err != nil {
					fmt.Println(err)
					continue
				}
				if err := hardcodeList(cmdsvc, ev.Command.UDID); err != nil {
					log.Println(err)
					continue
				}
				go func() {
					time.Sleep(10 * time.Second)
					pushsvc.Push(context.Background(), ev.Command.UDID)
				}()

			}
		}
	}()

	return nil
}

func hardcodeList(svc command.Service, udid string) error {
	ctx := context.Background()
	devInfo := &mdm.CommandRequest{
		RequestType: "DeviceInformation",
		UDID:        udid,
		Queries:     []string{"UDID"},
	}

	devConfigured := &mdm.CommandRequest{
		RequestType: "DeviceConfigured",
		UDID:        udid,
	}

	installProfile := &mdm.CommandRequest{
		RequestType: "InstallProfile",
		UDID:        udid,
		InstallProfile: mdm.InstallProfile{
			Payload: debugProfile,
		},
	}

	var requests = []*mdm.CommandRequest{
		devInfo,
		installProfile,
		devConfigured,
	}

	for _, r := range requests {
		_, err := svc.NewCommand(ctx, r)
		if err != nil {
			return err
		}
	}
	return nil
}

var debugProfile = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadDisplayName</key>
			<string>ManagedClient logging</string>
			<key>PayloadEnabled</key>
			<true/>
			<key>PayloadIdentifier</key>
			<string>com.apple.logging.ManagedClient.1</string>
			<key>PayloadType</key>
			<string>com.apple.system.logging</string>
			<key>PayloadUUID</key>
			<string>ED5DE307-A5FC-434F-AD88-187677F02222</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
			<key>Subsystems</key>
			<dict>
				<key>com.apple.ManagedClient</key>
				<dict>
					<key>DEFAULT-OPTIONS</key>
					<dict>
						<key>Default-Privacy-Setting</key>
						<string>Public</string>
						<key>Level</key>
						<dict>
							<key>Enable</key>
							<string>debug</string>
							<key>Persist</key>
							<string>debug</string>
						</dict>
					</dict>
				</dict>
			</dict>
		</dict>
		<dict>
			<key>PayloadDisplayName</key>
			<string>MDM debug mode</string>
			<key>PayloadType</key>
			<string>com.apple.mdmclient</string>
			<key>EnableDebug</key>
			<true/>
			<key>PayloadIdentifier</key>
			<string>com.apple.logging.ManagedClient.3</string>
			<key>PayloadUUID</key>
			<string>3EFF8784-7AE1-43E0-A2BA-6B77BBA54341</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
		</dict>
		<dict>
			<key>PayloadDisplayName</key>
			<string>ALR debug mode</string>
			<key>PayloadType</key>
			<string>com.apple.mcx.alr</string>
			<key>EnableDebug</key>
			<true/>
			<key>PayloadIdentifier</key>
			<string>com.apple.logging.ManagedClient.4</string>
			<key>PayloadUUID</key>
			<string>126C9C6B-AE28-4EA6-9BDB-FBB058A291B8</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
		</dict>
	</array>
	<key>PayloadDescription</key>
	<string>Enables ManagedClient debug mode and logging</string>
	<key>PayloadDisplayName</key>
	<string>MCX debug mode and logging</string>
	<key>PayloadIdentifier</key>
	<string>com.apple.logging.ManagedClient</string>
	<key>PayloadRemovalDisallowed</key>
	<false/>
	<key>PayloadScope</key>
	<string>System</string>
	<key>PayloadType</key>
	<string>Configuration</string>
	<key>PayloadUUID</key>
	<string>D30C25BD-E0C1-44C8-830A-964F27DAD4BA</string>
	<key>PayloadVersion</key>
	<integer>1</integer>
</dict>
</plist>`)
