CREATE TABLE "admin_type" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "admin_type" varchar UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "admin" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "username" varchar UNIQUE NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password" varchar NOT NULL,
  "active" boolean NOT NULL DEFAULT true,
  "type_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "last_login" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "user" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "username" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password" varchar NOT NULL,
  "telephone" int NOT NULL DEFAULT 0,
  "default_payment" bigint,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "user_session" (
  "id" uuid UNIQUE PRIMARY KEY NOT NULL,
  "user_id" bigint UNIQUE NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "expires_at" timestamptz NOT NULL
);

CREATE TABLE "address" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "address_line" varchar NOT NULL,
  "region" varchar NOT NULL,
  "city" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "user_address" (
  "user_id" bigint NOT NULL,
  "address_id" bigint UNIQUE NOT NULL,
  "default_address" bigint,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "user_review" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint UNIQUE NOT NULL,
  "ordered_product_id" bigint UNIQUE NOT NULL,
  "rating_value" int NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "payment_method" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint UNIQUE NOT NULL,
  "payment_type_id" int UNIQUE NOT NULL,
  "provider" varchar NOT NULL,
  "is_default" boolean NOT NULL DEFAULT false
);

CREATE TABLE "payment_type" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "value" varchar NOT NULL
);

CREATE TABLE "shopping_cart" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "shopping_cart_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "shopping_cart_id" bigint NOT NULL,
  "product_item_id" bigint UNIQUE NOT NULL,
  "qty" int NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "wish_list" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "wish_list_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "wish_list_id" bigint NOT NULL,
  "product_item_id" bigint UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "shop_order_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "product_item_id" bigint UNIQUE NOT NULL,
  "order_id" bigint NOT NULL,
  "quantity" int NOT NULL DEFAULT 0,
  "price" decimal NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "product_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "product_id" bigint NOT NULL,
  "product_sku" bigint NOT NULL,
  "qty_in_stock" int NOT NULL,
  "product_image" varchar NOT NULL,
  "price" decimal NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "product" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "category_id" bigint UNIQUE NOT NULL,
  "name" varchar NOT NULL,
  "description" varchar NOT NULL,
  "product_image" varchar NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "product_promotion" (
  "product_id" bigint UNIQUE NOT NULL,
  "promotion_id" bigint UNIQUE NOT NULL,
  "active" boolean NOT NULL DEFAULT false
);

CREATE TABLE "product_category" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "parent_category_id" bigint,
  "category_name" varchar NOT NULL
);

CREATE TABLE "promotion" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "name" varchar NOT NULL,
  "description" varchar NOT NULL,
  "discount_rate" int NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "start_date" timestamptz NOT NULL,
  "end_date" timestamptz NOT NULL
);

CREATE TABLE "category_promotion" (
  "category_id" bigint UNIQUE NOT NULL,
  "promotion_id" bigint UNIQUE NOT NULL,
  "active" boolean NOT NULL DEFAULT false
);

CREATE TABLE "variation" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "category_id" bigint UNIQUE NOT NULL,
  "name" varchar NOT NULL
);

CREATE TABLE "variation_option" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "variation_id" bigint UNIQUE NOT NULL,
  "value" varchar NOT NULL
);

CREATE TABLE "product_configuration" (
  "product_item_id" bigint UNIQUE NOT NULL,
  "variation_option_id" bigint UNIQUE NOT NULL
);

CREATE TABLE "shop_order" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint NOT NULL,
  "payment_method_id" bigint NOT NULL,
  "shipping_address_id" bigint NOT NULL,
  "order_total" decimal NOT NULL,
  "shipping_method_id" bigint NOT NULL,
  "order_status_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "order_status" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "status" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "shipping_method" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "name" varchar NOT NULL,
  "price" decimal NOT NULL
);

CREATE INDEX ON "user" ("username");

CREATE INDEX ON "user" ("email");

CREATE INDEX ON "user" ("telephone");

COMMENT ON COLUMN "payment_type"."value" IS 'for companies payment system like BCD';

COMMENT ON COLUMN "shop_order_item"."price" IS 'price of product when ordered';

COMMENT ON COLUMN "product_item"."product_image" IS 'may be used to show different images than original';

COMMENT ON COLUMN "product_item"."active" IS 'default is false';

COMMENT ON COLUMN "product"."active" IS 'default is false';

COMMENT ON COLUMN "product_promotion"."active" IS 'default is false';

COMMENT ON COLUMN "promotion"."active" IS 'default is false';

COMMENT ON COLUMN "category_promotion"."active" IS 'default is false';

COMMENT ON COLUMN "variation"."name" IS 'variation names like color, and size';

COMMENT ON COLUMN "variation_option"."value" IS 'variation values like Red, ans Size XL';

COMMENT ON COLUMN "order_status"."status" IS 'values like ordered, proccessed and delivered';

COMMENT ON COLUMN "shipping_method"."name" IS 'values like normal, or free';

ALTER TABLE "admin" ADD FOREIGN KEY ("type_id") REFERENCES "admin_type" ("id");

ALTER TABLE "user_session" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "user_address" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "user_address" ADD FOREIGN KEY ("address_id") REFERENCES "address" ("id");

ALTER TABLE "user_review" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "user_review" ADD FOREIGN KEY ("ordered_product_id") REFERENCES "shop_order_item" ("id");

ALTER TABLE "payment_method" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "payment_method" ADD FOREIGN KEY ("payment_type_id") REFERENCES "payment_type" ("id");

ALTER TABLE "shopping_cart" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "shopping_cart_item" ADD FOREIGN KEY ("shopping_cart_id") REFERENCES "shopping_cart" ("id");

ALTER TABLE "shopping_cart_item" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_item" ("id");

ALTER TABLE "wish_list" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "wish_list_item" ADD FOREIGN KEY ("wish_list_id") REFERENCES "wish_list" ("id");

ALTER TABLE "wish_list_item" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_item" ("id");

ALTER TABLE "shop_order_item" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_item" ("id");

ALTER TABLE "shop_order_item" ADD FOREIGN KEY ("order_id") REFERENCES "shop_order" ("id");

ALTER TABLE "product_item" ADD FOREIGN KEY ("product_id") REFERENCES "product" ("id");

ALTER TABLE "product" ADD FOREIGN KEY ("category_id") REFERENCES "product_category" ("id");

ALTER TABLE "product_promotion" ADD FOREIGN KEY ("product_id") REFERENCES "product" ("id");

ALTER TABLE "product_promotion" ADD FOREIGN KEY ("promotion_id") REFERENCES "promotion" ("id");

ALTER TABLE "product_category" ADD FOREIGN KEY ("parent_category_id") REFERENCES "product_category" ("id");

ALTER TABLE "category_promotion" ADD FOREIGN KEY ("category_id") REFERENCES "product_category" ("id");

ALTER TABLE "category_promotion" ADD FOREIGN KEY ("promotion_id") REFERENCES "promotion" ("id");

ALTER TABLE "variation" ADD FOREIGN KEY ("category_id") REFERENCES "product_category" ("id");

ALTER TABLE "variation_option" ADD FOREIGN KEY ("variation_id") REFERENCES "variation" ("id");

ALTER TABLE "product_configuration" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_item" ("id");

ALTER TABLE "product_configuration" ADD FOREIGN KEY ("variation_option_id") REFERENCES "variation_option" ("id");

ALTER TABLE "shop_order" ADD FOREIGN KEY ("shipping_address_id") REFERENCES "address" ("id");

ALTER TABLE "shop_order" ADD FOREIGN KEY ("shipping_method_id") REFERENCES "shipping_method" ("id");

ALTER TABLE "shop_order" ADD FOREIGN KEY ("order_status_id") REFERENCES "order_status" ("id");
