<?php
function check(string $date): bool
{
    if (!\DateTimeImmutable::createFromFormat('d/m/Y', $date)) {
        throw new \InvalidArgumentException('Use DD/MM/YYYY.');
    }
    $valid = false !== \DateTime::createFromFormat('Y-m-d', $date);
    $both = $valid && (\DateTime::createFromFormat('Y-m-d', $date));
    $kept = <warning descr="Reset omitted time fields when parsing a date-only value.">\DateTime::createFromFormat('Y-m-d', $date)</warning> == $date;
    return $both && $kept;
}
