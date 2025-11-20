CREATE STREAM cart_events (
    aggregate_id VARCHAR,
    event_type VARCHAR,
    event_id VARCHAR,
    timestamp BIGINT,
    version INT,
    data VARCHAR
)
WITH
    (
        kafka_topic = 'ec.cart-events',
        value_format = 'JSON',
        timestamp = 'timestamp'
    );

CREATE STREAM cart_events_structured AS
SELECT
    aggregate_id,
    event_type,
    event_id,
    timestamp,
    version,
    CASE
        WHEN event_type = 'CartCreatedEvent' THEN 'CREATED'
        WHEN event_type = 'CartItemAddedEvent' THEN 'ITEM_ADDED'
        WHEN event_type = 'CartSubmittedEvent' THEN 'SUBMITTED'
        WHEN event_type = 'CartPurchasedEvent' THEN 'PURCHASED'
        ELSE 'UNKNOWN'
    END as cart_action
FROM
    cart_events;

CREATE TABLE
    cart_state AS
SELECT
    aggregate_id,
    LATEST_BY_OFFSET (cart_action) as latest_action,
    LATEST_BY_OFFSET (timestamp) as last_activity_time,
    LATEST_BY_OFFSET (version) as current_version
FROM
    cart_events_structured
GROUP BY
    aggregate_id;

CREATE STREAM abandoned_carts AS
SELECT
    aggregate_id,
    latest_action,
    last_activity_time,
    (UNIX_TIMESTAMP () * 1000 - last_activity_time) / 1000 / 60 as minutes_since_activity
FROM
    cart_state
WHERE
    latest_action IN ('CREATED', 'ITEM_ADDED')
    AND (UNIX_TIMESTAMP () * 1000 - last_activity_time) > 1800000;

CREATE TABLE
    cart_funnel_metrics AS
SELECT
    COUNT(*) as total_carts,
    COUNT_DISTINCT (
        CASE
            WHEN latest_action = 'CREATED' THEN aggregate_id
        END
    ) as created_carts,
    COUNT_DISTINCT (
        CASE
            WHEN latest_action = 'ITEM_ADDED' THEN aggregate_id
        END
    ) as carts_with_items,
    COUNT_DISTINCT (
        CASE
            WHEN latest_action = 'SUBMITTED' THEN aggregate_id
        END
    ) as submitted_carts,
    COUNT_DISTINCT (
        CASE
            WHEN latest_action = 'PURCHASED' THEN aggregate_id
        END
    ) as purchased_carts
FROM
    cart_state;

-- Real-time cart activity monitoring
CREATE STREAM cart_activity_monitor AS
SELECT
    aggregate_id,
    cart_action,
    timestamp,
    'Cart ' + aggregate_id + ' ' + LCASE (cart_action) + ' at ' + TIMESTAMPTOSTRING (timestamp, 'yyyy-MM-dd HH:mm:ss') as activity_message
FROM
    cart_events_structured;