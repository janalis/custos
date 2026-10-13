<?php

function guardedByErrorCode(PDO $pdo, string $sql): string
{
    if (!$pdo->exec($sql)) {
        if ($pdo->errorCode() !== '00000') {
            throw new RuntimeException('failed: ' . $pdo->errorInfo()[2]);
        }
    }
    return 'ok';
}

function errorStateHeldInVariable(PDO $pdo, string $sql): bool
{
    if ($pdo->exec($sql) == false) {
        $info = $pdo->errorInfo();
        if ($info[0] !== '00000') {
            return false;
        }
    }
    return true;
}

function earlyReturnOnZeroRows(PDO $pdo, string $sql): int
{
    if (!($count = $pdo->exec($sql))) {
        if ($pdo->errorCode() === '00000') {
            return 0;
        }
        throw new RuntimeException('purge failed');
    }
    return $count;
}

function throwExpressionGuard(PDO $pdo, string $sql): void
{
    if (!$pdo->exec($sql)) {
        $pdo->errorCode() === '00000' || throw new RuntimeException('purge failed');
    }
}

function noRejection(PDO $pdo, string $sql): void
{
    if (!$pdo->exec($sql)) {
        echo 'nothing purged';
    }
}

function strictComparison(PDO $pdo, string $sql): void
{
    if ($pdo->exec($sql) === false) {
        throw new RuntimeException('purge failed');
    }
}
