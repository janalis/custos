<?php
// getLastErrors() runs only once the parse returned false.
function storedEarlyExit(string $s) {
    $date = DateTimeImmutable::createFromFormat('Y-m-d', $s);
    if ($date !== false) {
        return $date;
    }
    $errors = DateTimeImmutable::getLastErrors();
    return $errors['errors'];
}
function elseArm(string $s) {
    if ($date = DateTime::createFromFormat('Y-m-d', $s)) {
        return $date;
    } else {
        $errors = DateTime::getLastErrors();
        return $errors['error_count'];
    }
}
function elseifArm(string $s, bool $skip) {
    if ($skip) {
        return 0;
    } elseif (!date_create($s)) {
        $errors = DateTime::getLastErrors();
        return $errors['error_count'];
    }
    return 0;
}
function shortCircuit(string $s) {
    $date = DateTimeImmutable::createFromFormat('Y-m-d', $s) ?: throw new InvalidArgumentException(implode(', ', DateTimeImmutable::getLastErrors()['errors']));
    $other = date_create($s) or throw new InvalidArgumentException(DateTime::getLastErrors()['errors'][0]);
    return [$date, $other];
}
// Another parse, or no failed parse, before the call: still reported.
function parsedBetween(string $s, string $t) {
    $date = DateTimeImmutable::createFromFormat('Y-m-d', $s);
    if ($date !== false) {
        return $date;
    }
    $fallback = date_create($t);
    $errors = DateTimeImmutable::getLastErrors();
    return [$fallback, <warning descr="Handle the false result before reading parsing diagnostics.">$errors['errors']</warning>];
}
function parsedInLoop(string $s, array $more) {
    if (DateTime::createFromFormat('Y-m-d', $s)) {
        return [];
    }
    foreach ($more as $m) {
        $errors = DateTime::getLastErrors();
        echo <warning descr="Handle the false result before reading parsing diagnostics.">$errors['error_count']</warning>;
        $next = new DateTime($m);
    }
    return [];
}
function elseArmSucceeded(string $s) {
    if (!$date = DateTime::createFromFormat('Y-m-d', $s)) {
        return 0;
    } else {
        $errors = DateTime::getLastErrors();
        return <warning descr="Handle the false result before reading parsing diagnostics.">$errors['warning_count']</warning>;
    }
}
