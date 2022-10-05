# youtubetoolkit

## usage overview

When launching the CLI with a command for the first time, it will print a URL that you 
must open in your browser (if it fails to automagically open a browser for you) to authorize 
the CLI to act on your behalf on your YouTube account. 
The flag `--token` allows you to specify the filename where the CLI will store the auth token.

Data output will be printed to STDOUT, while info and errors to STDERR. Commands like `subscribe` 
and `playlist add` can receive data from STDIN. This allows you to do your wizardry with 
OS pipes and i/o redirections. 

For example, you can do something like this to copy subscriptions from one account to another:
```
$ youtubetoolkit -t account1 subscriptions list | youtubetoolkit -t account2 subscriptions add
```

Or you can add to a playlist the last 7 days video uploads from a list of channels:
```
$ youtubetoolkit lastuploads < channelIds.csv | youtubetoolkit playlists new test-playlist
```

Complete list of commands:
```
youtubetoolkit lastuploads --days <#>

youtubetoolkit subscriptions list
youtubetoolkit subscriptions add <channel id>
youtubetoolkit subscriptions del <channel id>

youtubetoolkit playlists
youtubetoolkit playlists new <playlist name>
youtubetoolkit playlists del <playlist id>

youtubetoolkit playlist --id <playlist_id>
youtubetoolkit playlist --id <playlist_id> add <video id>
```

Output formats available: `--csv`, `--table`, `--jsonl`.

Please use CLI flag `--help` to get additional help for every single command.

## install
```
go install ./cmd/youtubetoolkit
```

## requirement

To use the toolkit you need a OAuth2.0 client id and secret from Google Cloud:

- create a [new project](https://console.cloud.google.com/projectcreate)
- enable [YouTube Data API v3](https://console.cloud.google.com/apis/library/youtube.googleapis.com)
- configure the [OAuth consent screen](https://console.cloud.google.com/apis/credentials/consent)
    - for the sake of simplicity, set "External" type, don't publish the app (leave it in "testing"), and add your email to test users
- create a new [OAuth Client ID credential](https://console.cloud.google.com/apis/credentials) and download the JSON file when prompted
    - the application type must be "Desktop app"
- rename the json file to `client_secret.json` (it's the default filename the CLI will search for)
    - needless to say, do NOT disclose this file

As you may know, YouTube API access is subject to [quotas](https://console.cloud.google.com/apis/api/youtube.googleapis.com/quotas).
A GCloud dev project should have a total of 10000 units per day, plenty for read-only 
operations (only 1 unit) but not for creating/inserting things. 
API operations like subscribing to a channel, creating a playlist, adding a video to a 
playlist all cost 50 units. Or 200 inserts per day before running into "quota exceeded" errors. 
The CLI will output the quota impact after every execution.

## contributing

Pull requests are always welcome. 
