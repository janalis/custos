<?php
function check(string $date): bool
{
    if (!\DateTimeImmutable::createFromFormat('d/m/Y', $date)) {
        throw new \InvalidArgumentException('Use DD/MM/YYYY.');
    }
    $valid = false !== \DateTime::createFromFormat('Y-m-d', $date);
    $both = $valid && (\DateTime::createFromFormat('Y-m-d', $date));
    $kept = \DateTime::createFromFormat('!Y-m-d', $date) == $date;
    return $both && $kept;
}
