if [ -n "$DATABASE_URL" ]; then
  DB_ADDR=$(echo $DATABASE_URL | sed -e 's/^postgres:\/\///' -e 's/^postgresql:\/\///')
else
  echo "DATABASE_URL is not set"
  exit 1
fi

/nakama/nakama migrate up --database.address "$DB_ADDR"

exec /nakama/nakama --config /nakama/data/local.yml --database.address "$DB_ADDR"
