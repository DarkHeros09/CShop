ALTER TABLE "shopping_cart_item"
DROP CONSTRAINT shopping_cart_item_shopping_cart_id_product_item_id_key;

ALTER TABLE "payment_method"
DROP CONSTRAINT payment_method_user_id_payment_type_id_key;

DROP INDEX IF EXISTS notification_user_id_device_id_idx;
