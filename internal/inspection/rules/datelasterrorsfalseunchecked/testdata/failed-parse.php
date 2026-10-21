<?php
// A failed parse always records an error: getLastErrors() then returns an array.
function afterReturn(string $s) {
    if ($date = DateTimeImmutable::createFromFormat('Y-m-d', $s)) {
        return $date;
    }
    $errors = DateTimeImmutable::getLastErrors();
    throw new InvalidArgumentException(sprintf('%d errors', $errors['error_count']));
}
function notFalse(string $s) {
    if (false !== $date = DateTimeImmutable::createFromFormat('Y-m-d', $s)) {
        return $date;
    }
    $errors = DateTimeImmutable::getLastErrors();
    return $errors['errors'];
}
function notFalseRight(string $s) {
    if (($date = DateTimeImmutable::createFromFormat('Y-m-d', $s)) !== false) {
        return $date;
    }
    $errors = DateTimeImmutable::getLastErrors();
    return $errors['errors'];
}
// A thrown error may come before any parse (a TypeError): still reported.
function caught(string $s) {
    try {
        return new DateTimeImmutable($s);
    } catch (Throwable) {
        $errors = DateTimeImmutable::getLastErrors();
        return <warning descr="Handle the false result before reading parsing diagnostics.">$errors['error_count']</warning>;
    }
}
function negated(string $s) {
    if (!DateTime::createFromFormat('Y-m-d', $s)) {
        $errors = DateTime::getLastErrors();
        return $errors['errors'];
    }
    return [];
}
function strictFalse(string $s) {
    if (false === date_create_from_format('Y-m-d', $s)) {
        $errors = DateTime::getLastErrors();
        return $errors['error_count'];
    }
    if (date_create($s) === false) {
        $errors = DateTime::getLastErrors();
        return $errors['error_count'];
    }
    return 0;
}
// Not a failed parse, or another parse may run before the call.
function succeeded(string $s) {
    if ($date = DateTime::createFromFormat('Y-m-d', $s)) {
        $errors = DateTime::getLastErrors();
        return <warning descr="Handle the false result before reading parsing diagnostics.">$errors['warning_count']</warning>;
    }
    return 0;
}
function parsedAgain(string $s, string $t) {
    if (!DateTime::createFromFormat('Y-m-d', $s)) {
        $errors = DateTime::getLastErrors() and new DateTime($t);
        return <warning descr="Handle the false result before reading parsing diagnostics.">$errors['error_count']</warning>;
    }
    if (!DateTime::createFromFormat('Y-m-d', $t)) {
        $errors = DateTime::getLastErrors() and new Unresolved($t);
        return <warning descr="Handle the false result before reading parsing diagnostics.">$errors['error_count']</warning>;
    }
    return 0;
}
function fallsThrough(string $s) {
    if ($date = DateTime::createFromFormat('Y-m-d', $s)) {
        echo $date->format('c');
    }
    $errors = DateTime::getLastErrors();
    return <warning descr="Handle the false result before reading parsing diagnostics.">$errors['error_count']</warning>;
}
function notFirst(string $s) {
    if (!DateTime::createFromFormat('Y-m-d', $s)) {
        echo 'failed';
        $errors = DateTime::getLastErrors();
        return <warning descr="Handle the false result before reading parsing diagnostics.">$errors['error_count']</warning>;
    }
    return 0;
}
function otherGuard(bool $flag) {
    if (!$flag) {
        $errors = DateTime::getLastErrors();
        return <warning descr="Handle the false result before reading parsing diagnostics.">$errors['error_count']</warning>;
    }
    return 0;
}
