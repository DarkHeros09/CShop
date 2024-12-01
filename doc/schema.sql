-- SQL dump generated using DBML (dbml-lang.org)
-- Database: PostgreSQL
-- Generated at: 2024-11-30T22:16:55.194Z

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

CREATE TABLE "admin_session" (
  "id" uuid UNIQUE PRIMARY KEY NOT NULL,
  "admin_id" bigint NOT NULL,
  "refresh_token" varchar NOT NULL,
  "admin_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "expires_at" timestamptz NOT NULL
);

CREATE TABLE "user" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "username" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "password" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "is_email_verified" boolean NOT NULL DEFAULT false,
  "default_payment" bigint,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "verify_email" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint,
  "secret_code" varchar NOT NULL,
  "is_used" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '15 minutes')
);

CREATE TABLE "reset_passwords" (
  "id" bigserial PRIMARY KEY,
  "user_id" bigint NOT NULL,
  "email" varchar NOT NULL,
  "secret_code" varchar NOT NULL,
  "is_used" bool NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "expired_at" timestamptz NOT NULL DEFAULT (now() + interval '15 minutes')
);

CREATE TABLE "user_session" (
  "id" uuid UNIQUE PRIMARY KEY NOT NULL,
  "user_id" bigint NOT NULL,
  "refresh_token" varchar NOT NULL,
  "user_agent" varchar NOT NULL,
  "client_ip" varchar NOT NULL,
  "is_blocked" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "expires_at" timestamptz NOT NULL
);

CREATE TABLE "notification" (
  "user_id" bigint NOT NULL,
  "device_id" varchar,
  "fcm_token" varchar,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "app_policy" (
  "id" bigserial PRIMARY KEY,
  "policy" varchar,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "address" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "name" varchar NOT NULL,
  "telephone" int NOT NULL,
  "address_line" varchar NOT NULL,
  "region" varchar NOT NULL,
  "city" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "user_address" (
  "user_id" bigint NOT NULL,
  "address_id" bigint NOT NULL,
  "default_address" bigint,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "user_review" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint NOT NULL,
  "ordered_product_id" bigint UNIQUE NOT NULL,
  "rating_value" int NOT NULL DEFAULT 0,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "payment_method" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint NOT NULL,
  "payment_type_id" bigint NOT NULL,
  "provider" varchar NOT NULL,
  "is_default" boolean NOT NULL DEFAULT false
);

CREATE TABLE "payment_type" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "value" varchar UNIQUE NOT NULL,
  "is_active" boolean NOT NULL DEFAULT true
);

CREATE TABLE "shopping_cart" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "shopping_cart_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "shopping_cart_id" bigint NOT NULL,
  "product_item_id" bigint NOT NULL,
  "size_id" bigint NOT NULL,
  "qty" int NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "wish_list" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "user_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "wish_list_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "wish_list_id" bigint NOT NULL,
  "product_item_id" bigint NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "shop_order_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "product_item_id" bigint NOT NULL,
  "order_id" bigint NOT NULL,
  "quantity" int NOT NULL DEFAULT 0,
  "price" varchar NOT NULL,
  "discount" int NOT NULL DEFAULT 0,
  "shipping_method_price" varchar NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "featured_product_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "product_item_id" bigint NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "start_date" timestamptz NOT NULL DEFAULT 'now()',
  "end_date" timestamptz NOT NULL,
  "priority" int DEFAULT 0
);

CREATE TABLE "product_item" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "product_id" bigint NOT NULL,
  "image_id" bigint NOT NULL,
  "color_id" bigint NOT NULL,
  "product_sku" bigint NOT NULL,
  "price" varchar NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "product" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "category_id" bigint NOT NULL,
  "brand_id" bigint NOT NULL,
  "name" varchar NOT NULL,
  "description" varchar NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "product_size" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "product_item_id" bigint NOT NULL,
  "size_value" varchar NOT NULL,
  "qty" int NOT NULL DEFAULT 0
);

CREATE TABLE "product_color" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "color_value" varchar NOT NULL
);

CREATE TABLE "product_image" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "product_image_1" varchar NOT NULL,
  "product_image_2" varchar NOT NULL,
  "product_image_3" varchar NOT NULL
);

CREATE TABLE "product_promotion" (
  "product_id" bigint NOT NULL,
  "promotion_id" bigint NOT NULL,
  "product_promotion_image" varchar,
  "active" boolean NOT NULL DEFAULT false
);

CREATE TABLE "product_category" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "parent_category_id" bigint,
  "category_name" varchar UNIQUE NOT NULL,
  "category_image" varchar NOT NULL
);

CREATE TABLE "product_brand" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "brand_name" varchar UNIQUE NOT NULL,
  "brand_image" varchar NOT NULL
);

CREATE TABLE "home_page_text_banner" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "name" varchar UNIQUE NOT NULL,
  "description" varchar NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "promotion" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "name" varchar NOT NULL,
  "description" varchar NOT NULL,
  "discount_rate" bigint NOT NULL,
  "active" boolean NOT NULL DEFAULT false,
  "start_date" timestamptz NOT NULL,
  "end_date" timestamptz NOT NULL
);

CREATE TABLE "category_promotion" (
  "category_id" bigint UNIQUE NOT NULL,
  "promotion_id" bigint UNIQUE NOT NULL,
  "category_promotion_image" varchar,
  "active" boolean NOT NULL DEFAULT false
);

CREATE TABLE "brand_promotion" (
  "brand_id" bigint UNIQUE NOT NULL,
  "promotion_id" bigint UNIQUE NOT NULL,
  "brand_promotion_image" varchar,
  "active" boolean NOT NULL DEFAULT false
);

CREATE TABLE "variation" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "category_id" bigint NOT NULL,
  "name" varchar NOT NULL
);

CREATE TABLE "variation_option" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "variation_id" bigint,
  "value" varchar NOT NULL
);

CREATE TABLE "product_configuration" (
  "product_item_id" bigint NOT NULL,
  "variation_option_id" bigint NOT NULL
);

CREATE TABLE "shop_order" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "track_number" varchar NOT NULL,
  "order_number" int NOT NULL,
  "user_id" bigint NOT NULL,
  "payment_type_id" bigint NOT NULL,
  "shipping_address_id" bigint NOT NULL,
  "order_total" varchar NOT NULL,
  "shipping_method_id" bigint NOT NULL,
  "order_status_id" bigint,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "completed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "order_status" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "status" varchar UNIQUE NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT (now()),
  "updated_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "shipping_method" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "name" varchar UNIQUE NOT NULL,
  "price" varchar NOT NULL
);

CREATE INDEX ON "user" ("username");

CREATE INDEX ON "user" ("email");

CREATE UNIQUE INDEX ON "product_promotion" ("product_id", "promotion_id");

CREATE UNIQUE INDEX ON "category_promotion" ("category_id", "promotion_id");

CREATE UNIQUE INDEX ON "brand_promotion" ("brand_id", "promotion_id");

COMMENT ON COLUMN "payment_type"."value" IS 'for companies payment system like BCD';

COMMENT ON COLUMN "shop_order_item"."price" IS 'price of product when ordered';

COMMENT ON COLUMN "shop_order_item"."discount" IS 'discount of product when ordered';

COMMENT ON COLUMN "shop_order_item"."shipping_method_price" IS 'shipping method price when the order was made';

COMMENT ON COLUMN "product_item"."active" IS 'default is false';

COMMENT ON COLUMN "product"."active" IS 'default is false';

COMMENT ON COLUMN "product_promotion"."active" IS 'default is false';

COMMENT ON COLUMN "home_page_text_banner"."active" IS 'default is false';

COMMENT ON COLUMN "promotion"."active" IS 'default is false';

COMMENT ON COLUMN "category_promotion"."active" IS 'default is false';

COMMENT ON COLUMN "brand_promotion"."active" IS 'default is false';

COMMENT ON COLUMN "variation"."name" IS 'variation names like color, and size';

COMMENT ON COLUMN "variation_option"."value" IS 'variation values like Red, ans Size XL';

COMMENT ON COLUMN "order_status"."status" IS 'values like ordered, processed and delivered';

COMMENT ON COLUMN "shipping_method"."name" IS 'values like normal, or free';

ALTER TABLE "admin" ADD FOREIGN KEY ("type_id") REFERENCES "admin_type" ("id");

ALTER TABLE "admin_session" ADD FOREIGN KEY ("admin_id") REFERENCES "admin" ("id");

ALTER TABLE "verify_email" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id") ON DELETE SET NULL;

ALTER TABLE "reset_passwords" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "user_session" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

ALTER TABLE "notification" ADD FOREIGN KEY ("user_id") REFERENCES "user" ("id");

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

ALTER TABLE "featured_product_item" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_item" ("id");

ALTER TABLE "product_item" ADD FOREIGN KEY ("product_id") REFERENCES "product" ("id");

ALTER TABLE "product_item" ADD FOREIGN KEY ("image_id") REFERENCES "product_image" ("id");

ALTER TABLE "product_item" ADD FOREIGN KEY ("color_id") REFERENCES "product_color" ("id");

ALTER TABLE "product" ADD FOREIGN KEY ("category_id") REFERENCES "product_category" ("id");

ALTER TABLE "product" ADD FOREIGN KEY ("brand_id") REFERENCES "product_brand" ("id");

ALTER TABLE "product_size" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_item" ("id");

ALTER TABLE "product_promotion" ADD FOREIGN KEY ("product_id") REFERENCES "product" ("id");

ALTER TABLE "product_promotion" ADD FOREIGN KEY ("promotion_id") REFERENCES "promotion" ("id");

ALTER TABLE "product_category" ADD FOREIGN KEY ("parent_category_id") REFERENCES "product_category" ("id");

ALTER TABLE "category_promotion" ADD FOREIGN KEY ("category_id") REFERENCES "product_category" ("id");

ALTER TABLE "category_promotion" ADD FOREIGN KEY ("promotion_id") REFERENCES "promotion" ("id");

ALTER TABLE "brand_promotion" ADD FOREIGN KEY ("brand_id") REFERENCES "product_brand" ("id");

ALTER TABLE "brand_promotion" ADD FOREIGN KEY ("promotion_id") REFERENCES "promotion" ("id");

ALTER TABLE "variation" ADD FOREIGN KEY ("category_id") REFERENCES "product_category" ("id");

ALTER TABLE "variation_option" ADD FOREIGN KEY ("variation_id") REFERENCES "variation" ("id") ON DELETE SET NULL;

ALTER TABLE "product_configuration" ADD FOREIGN KEY ("product_item_id") REFERENCES "product_item" ("id");

ALTER TABLE "product_configuration" ADD FOREIGN KEY ("variation_option_id") REFERENCES "variation_option" ("id");

ALTER TABLE "shop_order" ADD FOREIGN KEY ("shipping_address_id") REFERENCES "address" ("id");

ALTER TABLE "shop_order" ADD FOREIGN KEY ("shipping_method_id") REFERENCES "shipping_method" ("id");

ALTER TABLE "shop_order" ADD FOREIGN KEY ("order_status_id") REFERENCES "order_status" ("id") ON DELETE SET NULL;
