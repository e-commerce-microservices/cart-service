CREATE TABLE  "cart" (
    "id" serial8 PRIMARY KEY,
    "customer_id" serial8  NOT NULL,
    "product_id" serial8 NOT NULL,
    "quantity" integer NOT NULL DEFAULT 1
);