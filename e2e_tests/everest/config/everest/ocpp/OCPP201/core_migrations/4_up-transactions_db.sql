CREATE TABLE TRANSACTIONS (
    TRANSACTION_ID TEXT NOT NULL UNIQUE,
    EVSE_ID INT NOT NULL UNIQUE,
    CONNECTOR_ID INT NOT NULL,
    TIME_START INT64 NOT NULL,
    SEQ_NO INT NOT NULL,
    CHARGING_STATE TEXT NOT NULL,
    ID_TAG_SENT INT NOT NULL
);