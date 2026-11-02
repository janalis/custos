<?php
function roles(array $list, $user) {
    $r = [];
    $r[] = in_array('admin', $list);
    $r[] = array_search(" editor ", $list);
    $r[] = in_array(\DateTime::class, $list);
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array('42', $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array(' ', $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array("$user", $list)</weak_warning>;
    $r[] = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array(7, $list)</weak_warning>;
    return $r;
}
