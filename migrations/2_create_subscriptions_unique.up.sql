CREATE UNIQUE INDEX subscriptions_unique_user_service_start
    ON subscriptions (user_id, service_name, start_date);
