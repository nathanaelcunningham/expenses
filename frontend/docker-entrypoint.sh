#!/bin/sh

# Replace the backend URL placeholder with the actual backend URL
if [ -n "$REACT_APP_API_URL" ]; then
    sed -i "s|BACKEND_URL_PLACEHOLDER|$REACT_APP_API_URL|g" /etc/nginx/nginx.conf
else
    echo "Warning: REACT_APP_API_URL not set, using placeholder"
fi

# Start nginx
exec nginx -g "daemon off;"