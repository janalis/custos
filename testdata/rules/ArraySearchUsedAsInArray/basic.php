<?php
function audit(array $roles, $who, $flag) {
    if (<warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles)</warning>) { log_hit(); }
    while (!<warning descr="Use 'in_array(...)' to test membership.">\array_search($who, $roles, true)</warning>) { $who = next_user(); }
    $ok = ((<warning descr="Use 'in_array(...)' to test membership.">array_search('root', $roles)</warning>)) && $flag;
    $tag = <warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles)</warning> ? 'known' : 'stranger';
    $any = $flag or <warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles)</warning>;
    do { $who = next_user(); } while (<warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles)</warning>);
    if ($flag) {} elseif ((<warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles)</warning>)) {}

    $miss = <warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles) === FALSE</warning>;
    $hit  = <warning descr="Use 'in_array(...)' to test membership.">false !== array_search($who, $roles)</warning>;
    $miss2 = <warning descr="Use 'in_array(...)' to test membership.">\false === \array_search($who, $roles)</warning>;
    $odd  = array_search($who, $roles) === <error descr="array_search() cannot return true; this comparison never changes.">true</error>;
    $odd2 = <error descr="array_search() cannot return true; this comparison never changes.">TRUE</error> !== array_search($who, $roles);
    return [$ok, $tag, $any, $miss, $hit, $miss2, $odd, $odd2];
}
