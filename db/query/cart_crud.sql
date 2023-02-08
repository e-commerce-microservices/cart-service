-- name: CreateCart :exec
INSERT INTO "cart" (
    "customer_id", "product_id"
) VALUES (
    $1, $2
);

-- name: DeleteCart :exec
DELETE FROM "cart"
WHERE "id" = $1 AND "customer_id" = $2;