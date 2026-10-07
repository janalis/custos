<?php
function granted(string $who, array $roles, array $admins): bool {
    if (<warning descr="Use 'in_array(...)' to test membership.">ARRAY_SEARCH($who, $roles)</warning>) {
        return true;
    }
    $denied = <warning descr="Use 'in_array(...)' to test membership.">\Array_Search($who, $admins) === false</warning>;
    $never = Array_Search($who, $admins) !== <error descr="array_search() cannot return true; this comparison never changes.">true</error>;
    $key = Array_Search($who, $admins);
    return !$denied && $never && $key;
}
