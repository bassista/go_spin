#!/bin/sh
set -e

# Default UID/GID to 1000 if not provided
UID=${UID:-1000}
GID=${GID:-1000}

GROUPNAME=appgroup
USERNAME=appuser

# Create group if no group with requested GID exists
if grep -qE "^[^:]+:[^:]*:${GID}:" /etc/group; then
  GROUPNAME=$(awk -F: -v gid="$GID" '$3==gid{print $1; exit}' /etc/group)
else
  addgroup -g "$GID" "$GROUPNAME"
fi

# Create user if no user with requested UID exists
if grep -qE "^[^:]+:[^:]*:${UID}:" /etc/passwd; then
  USERNAME=$(awk -F: -v uid="$UID" '$3==uid{print $1; exit}' /etc/passwd)
else
  adduser -D -u "$UID" -G "$GROUPNAME" -h /home/"$USERNAME" -s /bin/sh "$USERNAME"
fi

# Ensure ownership of application directory
chown -R "${UID}:${GID}" /app

# Exec the requested command as the chosen user:group.
# Prefer `gosu`, then `su-exec`, else fall back to `su`.
if command -v gosu >/dev/null 2>&1; then
  exec gosu "$USERNAME":"$GROUPNAME" "$@"
elif command -v su-exec >/dev/null 2>&1; then
  exec su-exec "$USERNAME":"$GROUPNAME" "$@"
else
  # Fallback: use su to run command (joins args into a single command string)
  CMD_STR="$*"
  exec su -s /bin/sh "$USERNAME" -c "$CMD_STR"
fi
