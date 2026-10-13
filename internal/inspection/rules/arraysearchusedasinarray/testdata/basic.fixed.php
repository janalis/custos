<?php
function audit(array $roles, $who, $flag) {
    if (in_array($who, $roles)) { log_hit(); }
    while (!\in_array($who, $roles, true)) { $who = next_user(); }
    $ok = ((in_array('root', $roles))) && $flag;
    $tag = in_array($who, $roles) ? 'known' : 'stranger';
    $any = $flag or in_array($who, $roles);
    do { $who = next_user(); } while (in_array($who, $roles));
    if ($flag) {} elseif ((in_array($who, $roles))) {}

    $miss = !in_array($who, $roles);
    $hit  = in_array($who, $roles);
    $miss2 = !\in_array($who, $roles);
    $odd  = array_search($who, $roles) === true;
    $odd2 = TRUE !== array_search($who, $roles);
    return [$ok, $tag, $any, $miss, $hit, $miss2, $odd, $odd2];
}
