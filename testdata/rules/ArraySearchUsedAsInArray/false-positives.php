<?php
function audit(array $roles, $who, $flag, $obj) {
    $key   = array_search($who, $roles) ?: 'none';
    $other = $flag ?? array_search($who, $roles);
    $br    = $flag ? array_search($who, $roles) : 1;
    $loose = array_search($who, $roles) == false;
    $neq   = array_search($who, $roles) != true;
    $zero  = array_search($who, $roles) === 0;
    $null  = array_search($who, $roles) !== null;
    $wrap  = (array_search($who, $roles)) !== false;
    $one   = array_search($who);
    $spr   = array_search(...$roles);
    for ($i = 0; array_search($i, $roles); $i++) {}
    $k = array_search($who, $roles);
    use_it(array_search($who, $roles));
    if ($obj->array_search($who, $roles)) {}
    if (Foo::array_search($who, $roles)) {}
    return $flag xor array_search($who, $roles);
}
