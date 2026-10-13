<?php
function conditionalEnable(bool $emulate) {
    $pdo = new PDO('mysql:host=localhost;dbname=sample', 'u', 'p', [PDO::ATTR_EMULATE_PREPARES => false]);
    if ($emulate) {
        $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, true);
    }
    return $emulate ? $pdo->prepare('SELECT :id + :id') : $pdo->prepare('SELECT :a + :b');
}

function conditionalDirectQuery(bool $direct) {
    $pdo = new PDO('mysql:host=localhost;dbname=sample');
    $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false);
    if ($direct) {
        $pdo->setAttribute(PDO::MYSQL_ATTR_DIRECT_QUERY, true);
    }
    return $pdo->prepare('SELECT :id + :id');
}

function attributeChosenAtRunTime(int $attribute) {
    $pdo = new PDO('mysql:host=localhost;dbname=sample', 'u', 'p', [PDO::ATTR_EMULATE_PREPARES => false]);
    $pdo->setAttribute($attribute, true);
    return $pdo->prepare('SELECT :id + :id');
}

function enabledThroughClosure() {
    $pdo = new PDO('mysql:host=localhost;dbname=sample', 'u', 'p', [PDO::ATTR_EMULATE_PREPARES => false]);
    $enable = function () use ($pdo) {
        $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, true);
    };
    $enable();
    return $pdo->prepare('SELECT :id + :id');
}

function enabledForTheNextIteration(array $rows) {
    $pdo = new PDO('mysql:host=localhost;dbname=sample', 'u', 'p', [PDO::ATTR_EMULATE_PREPARES => false]);
    foreach ($rows as $row) {
        $pdo->prepare('SELECT :id + :id');
        if ($row) {
            $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, true);
        }
    }
}
