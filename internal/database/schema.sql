CREATE TABLE Events (
    event_id TEXT PRIMARY KEY NOT NULL DEFAULT (
        lower(
            hex(randomblob(4)) || '-' ||
            hex(randomblob(2)) || '-' ||
            '4' || substr(hex(randomblob(2)), 2) || '-' ||
            substr('89ab', abs(random()) % 4 + 1, 1) ||
            substr(hex(randomblob(2)), 2) || '-' ||
            hex(randomblob(6))
        )
    ),
    time TIMESTAMP NOT NULL,
    proof INTEGER
);

CREATE TABLE Swinces (
    event_id TEXT NOT NULL,
    swince_id TEXT NOT NULL DEFAULT (
        lower(
            hex(randomblob(4)) || '-' ||
            hex(randomblob(2)) || '-' ||
            '4' || substr(hex(randomblob(2)), 2) || '-' ||
            substr('89ab', abs(random()) % 4 + 1, 1) ||
            substr(hex(randomblob(2)), 2) || '-' ||
            hex(randomblob(6))
        )
    ),
    participant_id TEXT NOT NULL,
    nominee_id TEXT,
    fulfillment_id TEXT,
    PRIMARY KEY (event_id, swince_id),
    FOREIGN KEY (event_id) REFERENCES Events(event_id) ON DELETE CASCADE,
    FOREIGN KEY (fulfillment_id) REFERENCES Swinces(swince_id)
);

CREATE TABLE Rulesets (
    ruleset_id TEXT PRIMARY KEY NOT NULL DEFAULT (
        lower(
            hex(randomblob(4)) || '-' ||
            hex(randomblob(2)) || '-' ||
            '4' || substr(hex(randomblob(2)), 2) || '-' ||
            substr('89ab', abs(random()) % 4 + 1, 1) ||
            substr(hex(randomblob(2)), 2) || '-' ||
            hex(randomblob(6))
        )
    ),
    start_time TIMESTAMP NOT NULL
);

