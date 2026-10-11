<?php

function messageFromErrorInfo(PDO $pdo, string $sql): void
{
    if (!<warning descr="Compare database execution failure strictly with false.">$pdo->exec($sql)</warning>) {
        throw new RuntimeException('failed: ' . $pdo->errorInfo()[2]);
    }
}

function errorInfoThenUnconditionalThrow(PDO $pdo, string $sql): void
{
    if (<warning descr="Compare database execution failure strictly with false.">$pdo->exec($sql)</warning> == false) {
        $info = $pdo->errorInfo();
        echo 'purge failed';
        throw new RuntimeException($info[2]);
    }
}

function otherConnectionConsulted(PDO $pdo, PDO $replica, string $sql): bool
{
    if (!<warning descr="Compare database execution failure strictly with false.">$pdo->exec($sql)</warning>) {
        if ($replica->errorCode() !== '00000') {
            return false;
        }
    }
    return true;
}

function unrelatedGuard(PDO $pdo, string $sql, bool $strict): void
{
    if (!<warning descr="Compare database execution failure strictly with false.">$pdo->exec($sql)</warning>) {
        if ($strict) {
            throw new RuntimeException('nothing purged');
        }
    }
}

function singleStatementBranch(PDO $pdo, string $sql): bool
{
    if (!<warning descr="Compare database execution failure strictly with false.">$pdo->exec($sql)</warning>)
        return false;
    return true;
}
