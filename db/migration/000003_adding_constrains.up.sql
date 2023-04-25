ALTER TABLE "shopping_cart_item"
ADD UNIQUE (shopping_cart_id, product_item_id);