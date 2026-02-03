#!/bin/sh
set -e

# Default UID/GID to 1000 if not provided
UID=${UID:-1000}
GID=${GID:-1000}

GROUPNAME=appgroup
USERNAME=appuser

# Detect Alpine vs Debian-based image
is_alpine=false
if [ -f /etc/alpine-release ] || (grep -qi alpine /etc/os-release 2>/dev/null); then
  is_alpine=true
fi

# Create group if no group with requested GID exists
EXISTING_GROUP=$(awk -F: -v gid="$GID" '$3==gid{print $1; exit}' /etc/group || true)
if [ -n "$EXISTING_GROUP" ]; then
  GROUPNAME="$EXISTING_GROUP"
else
  if [ "$is_alpine" = "true" ]; then
    addgroup -g "$GID" "$GROUPNAME"
  else
    groupadd -g "$GID" "$GROUPNAME" || true
  fi
fi

# Create user if no user with requested UID exists
EXISTING_USER=$(awk -F: -v uid="$UID" '$3==uid{print $1; exit}' /etc/passwd || true)
if [ -n "$EXISTING_USER" ]; then
  USERNAME="$EXISTING_USER"
else
  if [ "$is_alpine" = "true" ]; then
    adduser -D -u "$UID" -G "$GROUPNAME" -h /home/"$USERNAME" -s /bin/sh "$USERNAME"
  else
    useradd -m -u "$UID" -g "$GROUPNAME" -d /home/"$USERNAME" -s /bin/sh "$USERNAME" || true
  fi
fi

# Ensure ownership of application directory and home
chown -R "${UID}:${GID}" /app 2>/dev/null || true
chown -R "${UID}:${GID}" /home/"$USERNAME" 2>/dev/null || true

# Exec the requested command as the chosen user:group.
# Prefer `gosu`, then `su-exec`, else fall back to `su`.
if command -v gosu >/dev/null 2>&1; then
  exec gosu "$USERNAME":"$GROUPNAME" "$@"
elif command -v su-exec >/dev/null 2>&1; then
  exec su-exec "$USERNAME":"$GROUPNAME" "$@"
else
  CMD_STR="$*"
  exec su -s /bin/sh "$USERNAME" -c "$CMD_STR"
fi
