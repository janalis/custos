<?php
function repeatedAfterOtherAttribute(bool $strict) {
    $pdo = new PDO('mysql:host=localhost;dbname=sample', 'u', 'p', [PDO::ATTR_EMULATE_PREPARES => false]);
    if ($strict) {
        $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
    }
    return <warning descr="Use a distinct marker for each native parameter.">$pdo->prepare('SELECT :id + :id')</warning>;
}

function repeatedAfterRestoredProof(bool $emulate) {
    $pdo = new PDO('mysql:host=localhost;dbname=sample');
    if ($emulate) {
        $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, true);
    }
    $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, false);
    return <warning descr="Use a distinct marker for each native parameter.">$pdo->prepare('SELECT :id + :id')</warning>;
}

function repeatedBeforeLaterEnable() {
    $pdo = new PDO('mysql:host=localhost;dbname=sample', 'u', 'p', [PDO::ATTR_EMULATE_PREPARES => false]);
    $statement = <warning descr="Use a distinct marker for each native parameter.">$pdo->prepare('SELECT :id + :id')</warning>;
    $pdo->setAttribute(PDO::ATTR_EMULATE_PREPARES, true);
    return $statement;
}
