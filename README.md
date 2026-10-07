## This is an e-Mail fetch & store bot/tool written in GoLang for AI Agents to connect

# What does it do?
1. It take user credentials from web portal, stores them in a sqlite database for connections.
2. It supports IMAP(SSL/TLS Supported), OAuth2 (GMail, Yahoo Mail, MS Outlook Mail and etc.)
3. It downloads the mail contents and stores in a local file system separated by users.
4. All attachments are stored separately with original file names with original file formats, separated by users.
5. It provides secure API with API tokens (set validation period) to allow AI agents to access.
6. It provides manager panel to control entire system, see system telemetries.

# Who can use?
1. Anyone who wants to let AI Agents like Muse.AI and your own bots to process your emails.
2. Anyone who wants to get the emails via IMs like Matrix, WeChat, Telegram and etc without email client.
3. Any developpers who are interested in.

# What to expect?
1. It works, but do not quarrenty error-free & bug-free. It is still at young stage.
2. Put this in a docker, and use Nginx for reverse proxy. Exposing the ports directly is NOT RECOMMENDED.
3. Currently, the API only supports fetching emails, and DOES NOT SUPPORT marking, writing, deleting, moving emails in your original server(s)。
4. So clean and organize and clean your email boxes regularly to avoid mail box getting full.
5. I will try to continuously improve the application and add more features based on user feedback.
6. I will try to add email deletion and management features via API in future updates, with different branch as it may add even more serious damages to your emails if not handled properly.

# Security & Important Things
1. Put this aplication in a docker, and use Nginx for reverse proxy. I am telling you again!!
2. Exposing the ports directly is like going out naked. This application does not provide HTTPS by itself, so use Nginx for HTTPS/SSL Certs.
3. The attachments can fill your drive quickly especially when you share the server with other people (like your family or colleages and friends). Get a large disk for the file storage.
4. If you use this application which is hosted by someone else, he/she/they can see your emails and attachments if the server file system is accessed. So, use this application when only trusted server is used.
5. Be cautious with API tokens, as they grant access to your emails and attachments. Keep them secure and rotate them regularly.
6. The default validation period for API tokens are 90 days. Keep it shorter than 90 days if possible.
7. Keep your portal credentials secure and do not share them with anyone. Once compromised, your emails and attachments can be accessed by unauthorized parties via new API tokens or direct access to the server.
8. Regularly monitor and audit the usage of API tokens to detect any unauthorized access or suspicious activity.
9. Always keep backups of your emails and attachments, as this application does not provide guaranteed backup mechanism by itself.
10. Downloaded emails and attachments to the application server can be lost if the local file system is compromised or deleted.