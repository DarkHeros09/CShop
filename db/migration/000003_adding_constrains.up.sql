ALTER TABLE "shopping_cart_item"
ADD UNIQUE (shopping_cart_id, product_item_id);

ALTER TABLE "payment_method"
ADD UNIQUE (user_id, payment_type_id);

CREATE UNIQUE INDEX notification_user_id_device_id_idx
ON "notification" (user_id,device_id)
WHERE device_id IS NOT NULL;