# Snapshotting

This project contains a tool to create and restore user-space snapshots.

It is derived from the project [s3backup](https://github.com/parlaynu/s3backup), rearranging the core operators
and algorithm to work for user-space snapshots instead of backups.

The snapshot is controlled through a configuration file, with an example shown [here](config/config.yaml).
The default location for the configuration file is `~/.config/snappit/config.yaml`; this can be overridden
on the command line.

## Concepts

Arena - this is the area that contains the snapshots. There are multiple snapshots in an arena.
The arena is defined in the configuration file.

Snapshot - this contains one or more archives as defined in the configuration file. The snapshot
name is a directory within the arena.

Archive - this is the basic unit that is snapshotted. The archives are defined in the config file. 
It consists of a manifest file with the full list of files included in the snapshot
along with each file's metadata such as size, modify time and, importantly, the content hash.

Arena data - the actual files are stored in the data location within the arena in files named
by the content hash of the file. This provides deduplication by default. The content hash in the
manifest file is used to locate the file within the data directory.

## Quick Start

For a quick demonstration of it's use, take the example configuration file and customise it to create
some snapshots.

To create the baseline snapshot, run:

    $ snappit create baseline

Then modify some files in the source directories and create a new snapshot:

    $ snappit create v001

To list the snapshots available, run the list command:

    $ snappit list

To restore to the baseline snapshot:

    $ snappit restore baseline

This last command will restore the source directories to how they were when the baseline
snapshot was created.

