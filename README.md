## This is an e-Mail fetch & store bot/tool written in GoLang for AI Agents to connect

# What does it do?
1. It take user credentials from web portal, stores them in a sqlite database for connections.
2. It supports IMAP(SSL/TLS Supported), OAuth2 (GMail, Yahoo Mail, MS Outlook Mail and etc.)
3. It downloads the mail contents and store in a local file system separated by users.
4. All attachments are stored separately with original file name as files, separated by users.
5. It provides secure API with API token (set validation period) to allow AI agents to access.
6. It provides manager panel to control entire system, see system telemetries.

# Who can use?
1. Anyone who wants to let AI Agents like Muse.AI and your own bots to process your emails.
2. Anyone who wants to get the emails via IMs like Matrix, WeChat, Telegram and etc without email client.
3. Any developpers who are interested in.

# What to expect?
1. It works, but do not quarrenty error-free & bug-free. It is still at young stage.
2. Put this in a docker, and use Nginx for reverse proxy. Exposing the ports directly is NOT RECOMMENDED.

# Security & Important Things
1. Put this aplication in a docker, and use Nginx for reverse proxy. I am telling you again!!
2. Exposing the ports directly is like going out naked. The app does not provide HTTPS by itself, so use Nginx for HTTPS/SSL Certs.
3. The attachments can fill your drive quickly especially you share the server with other people (like your family or colleges and friends). Get a large disk for the file storage.
4. If you use this application which hosted by someone else, he/she/they can see your emails and attachments if the server file system is accessed. So, use this application when only trusted server is used.