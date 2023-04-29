CREATE TABLE events (
    ID          TEXT PRIMARY KEY,
    Number       int,
    MQTT         VARCHAR(255),
    InventoryID  VARCHAR(255),
    UnitGUID     VARCHAR(255),
    MessageID    VARCHAR(255),
    MessageText  TEXT,
    Context      VARCHAR(255),
    MessageClass VARCHAR(255), -- TODO: ADD CHECK,
    Level        INTEGER,
    Area         VARCHAR(255),
    Address      VARCHAR(255),
    Block        BOOLEAN,
    Type         VARCHAR(255),
    Bit          INTEGER,
    InvertBit    INTEGER
);
CREATE TABLE files (
    name VARCHAR(255) PRIMARY KEY,
    error VARCHAR(255)
);
CREATE TABLE relations(
    file_name VARCHAR(255),
    event_id TEXT,
    FOREIGN KEY(file_name) REFERENCES files(name),
    FOREIGN KEY(event_id) REFERENCES events(ID)
)