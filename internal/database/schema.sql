CREATE TABLE Swinces (
  swince_id UUID NOT NULL UNIQUE PRIMARY KEY DEFAULT (lower(hex(randomblob(4)) || '-' || hex(randomblob(2)) || '-' || '4' || substr(hex(randomblob(2)),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(hex(randomblob(2)),2) || '-' || hex(randomblob(6)))),
  media BLOB NOT NULL UNIQUE,
  uploaded_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE Swinceurs (
  swince_id UUID NOT NULL,
  discord_id INTEGER NOT NULL, -- is a 64 bit int
  late_swince_tax REAL NOT NULL CHECK (late_swince_tax BETWEEN 0 AND 1),
  PRIMARY KEY (task_id, tag_id),
  FOREIGN KEY (swince_id) REFERENCES Swinces(swince_id) ON DELETE CASCADE
);

CREATE TABLE Nominees (
  swince_id UUID NOT NULL,
  discord_id INTEGER NOT NULL, -- is a 64 bit int
  PRIMARY KEY (task_id, tag_id),
  FOREIGN KEY (swince_id) REFERENCES Swinces(swince_id) ON DELETE CASCADE
);

