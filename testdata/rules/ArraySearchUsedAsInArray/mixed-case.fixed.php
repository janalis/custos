<?php
function granted(string $who, array $roles, array $admins): bool {
    if (in_array($who, $roles)) {
        return true;
    }
    $denied = !\in_array($who, $admins);
    $never = Array_Search($who, $admins) !== true;
    $key = Array_Search($who, $admins);
    return !$denied && $never && $key;
}
