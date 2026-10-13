<?php
class Account {}

function check($other, Account $current, $flag) {
    // identity check against another object: an ordinary condition
    if ($other instanceof Account && $other !== $current) {}
    if ($other instanceof Account && $current != $other) {}
    // comparisons with literals next to instanceof stay reported
    if (<weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$other !== null</weak_warning> && $other instanceof Account) {}
    if ($other instanceof Account && <weak_warning descr="Equality check on a value also tested with instanceof; verify the logic.">$other == 'admin'</weak_warning>) {}
}
