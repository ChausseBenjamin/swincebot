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
    time TIMESTAMP NOT NULL, -- submission time (should default to now)
    proof INTEGER -- discord messageID of the video on swince channel
);

CREATE TABLE Swinces (
    event_id TEXT NOT NULL, -- video in which the swince was performed (multiple swinces during a single event possible)
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
    participant_id INTEGER NOT NULL, -- person performing the swince (Discord user ID)
    nominee_id INTEGER, -- person nominated: optional as you *can* nominate nobody (loser)
    fulfillment_id TEXT, -- ID of the Swince that fullfills the nomination
    PRIMARY KEY (event_id, swince_id),
    FOREIGN KEY (event_id) REFERENCES Events(event_id) ON DELETE CASCADE,
    FOREIGN KEY (fulfillment_id) REFERENCES Swinces(swince_id)
);

CREATE TABLE Seasons (
    start_time TIMESTAMP PRIMARY KEY NOT NULL,
    ruleset TEXT NOT NULL
);

