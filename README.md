# ws-chat-app

A small real-time chat app. Create an account, pick a room by name, and chat with anyone else in that room. Messages appear instantly, you can see when someone is typing, and the conversation is saved so it's still there when you come back.

It's made of three pieces that all run on your own computer:

- a **PostgreSQL database** (run with Docker) that stores accounts and messages
- a **Go server** that handles logins and passes messages between people
- a **Next.js website** that you open in your browser

## What you'll need

- [Docker Desktop](https://www.docker.com/products/docker-desktop/), for the database
- [Go](https://go.dev/dl/) 1.25 or newer
- [Node.js](https://nodejs.org/) 20.9 or newer, which includes `npm`

## Getting started

Run these commands from the project folder. You'll end up with two terminal windows open, one for the server and one for the website.

### 1. Start the database

Make sure Docker Desktop is running, then:

```bash
docker compose up -d
```

This starts PostgreSQL in the background. Your data is kept between restarts.

### 2. Start the server

```bash
cd backend
go run ./cmd/server
```

The first time it starts, the server creates the database tables it needs, and it adds any new ones automatically after updates. When you see `Server listening on http://localhost:8080 (development)`, it's ready. Leave this terminal open.

You'll also see a warning that `JWT_SECRET` isn't set. That's fine on your own computer. See [Settings](#settings) if you want to set one.

### 3. Start the website

In a second terminal:

```bash
cd frontend
npm install
npm run dev
```

Then open **http://localhost:3000** in your browser.

## Using the app

1. On the home page, click **New user? Register** and choose a username and a password (at least 6 characters).
2. After registering you're taken to the login page. Log in with the details you just chose.
3. Type the name of a room, such as `general`, `games` or anything else, and click **Join room**. A room is created the first time someone joins it.
4. Start chatting.

To try it with more than one person, open the app in a second browser or a private window, log in as a different user, and join the same room.

At the top of each room you'll see who you're logged in as, along with links to **Switch room** and **Log out**.

### Good to know

- **Guests can read but not post.** The "go straight to chat lobby" link on the home page lets you join any room without an account. You'll see its recent history and new messages as they arrive, but you need to log in to send anything.
- **History:** opening a room shows its 50 most recent messages.
- **Messages can be up to 1,000 characters.** A counter appears as you get close to the limit.
- **If the connection drops**, the status above the chat changes to `reconnecting` and the app keeps retrying on its own. Anything you typed stays in the message box, and messages sent while you were away appear once you're back.
- **Logins last 24 hours.** Logging out ends that login straight away, even if the app is open in other tabs. If your login ends while you're in a room, the room tells you and switches you to guest mode, so log in again to keep chatting.

## Settings

You can usually skip this section, because the defaults work on your own computer. Change these only when you need something different, such as other ports or running the app somewhere other than `localhost`.

### Server

The easiest way to change settings is with a `.env` file in the project folder. Copy the example and edit it:

```bash
cp .env.example .env
```

Both the server and Docker read this file. It's ignored by git, so passwords and secrets you put in it stay on your computer. Restart the server after changing it. When the file is found, the server prints `loaded settings from ...` as it starts.

These are the settings the server understands:

| Variable          | What it does                                                        | Default                                                               |
| ----------------- | ------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `APP_ENV`         | `development` or `production` (see below)                           | `development`                                                         |
| `PORT`            | The port the server listens on                                      | `8080`                                                                |
| `DATABASE_URL`    | Which PostgreSQL database to use                                    | `postgres://chatuser:chatpass@localhost:5432/chatapp?sslmode=disable` |
| `JWT_SECRET`      | The secret used to sign logins                                      | an insecure development value                                         |
| `ALLOWED_ORIGINS` | Website addresses allowed to talk to the server, separated by commas | `http://localhost:3000`                                               |

`.env.example` explains each one. It also has `POSTGRES_USER`, `POSTGRES_PASSWORD` and `POSTGRES_DB`, which set up the Docker database. If you change those, update `DATABASE_URL` to match. Docker only applies them when it first creates the database, so changing them later means starting with a fresh one (`docker compose down -v`, which erases existing accounts and messages).

You can also set any of these as a regular environment variable, which takes priority over `.env`. That's handy for a one-off change, or on a host that manages its own settings:

```bash
# macOS, Linux or Git Bash
PORT=9000 go run ./cmd/server
```

```powershell
# PowerShell
$env:PORT = "9000"; go run ./cmd/server
```

The server looks for `.env` in the folder you start it from, and then in each folder above that up to the project folder. To use a file somewhere else, set `ENV_FILE` to its path.

If you'd rather use a PostgreSQL install you already have instead of Docker, point `DATABASE_URL` at it. The server will create its tables there.

### Website

If you change where the server runs, tell the website by copying `frontend/.env.example` to `frontend/.env.local` and editing it:

```bash
NEXT_PUBLIC_API_BASE_URL=http://localhost:9000
```

Restart `npm run dev` afterwards. If you change the website's own address or port, add the new address to the server's `ALLOWED_ORIGINS`, or the server will turn the website's requests away.

### Production mode

With `APP_ENV=production`, the server refuses to start unless `DATABASE_URL`, `ALLOWED_ORIGINS` and a `JWT_SECRET` of at least 32 random characters are all set. You can generate a secret with `openssl rand -hex 32`. The default database password is only meant for local use, so change `POSTGRES_PASSWORD` too, along with the password in `DATABASE_URL`.

To run an optimised build of the website instead of the development server:

```bash
cd frontend
npm run build
npm start
```

`NEXT_PUBLIC_API_BASE_URL` is fixed when you run `npm run build`, so set it before building.

## Running the tests

The website's tests run on their own:

```bash
cd frontend
npm test
```

The server's tests need a PostgreSQL database to work in. They create and remove their own temporary tables, so it's safe to use the Docker database from step 1. Uncomment the `TEST_DATABASE_URL` line in your `.env` (it's already filled in), then run:

```bash
cd backend
go test ./...
```

Without `TEST_DATABASE_URL`, only the tests that don't need a database run, and the rest are skipped.

These tests and a production build also run automatically on GitHub for every push to `main` or `dev` and for every pull request.

## Stopping everything

Press `Ctrl+C` in the server and website terminals, then stop the database:

```bash
docker compose down
```

Your accounts and messages are kept. To erase them as well, run `docker compose down -v`.
