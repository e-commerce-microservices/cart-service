-- name: CreateCart :exec
INSERT INTO "cart" (
    "customer_id", "product_id", "quantity"
) VALUES (
    $1, $2, $3
);

-- name: DeleteCart :exec
DELETE FROM "cart"
WHERE "id" = $1 AND "customer_id" = $2;

-- name: GetCart :many
SELECT * FROM "cart"
WHERE "customer_id" = $1;