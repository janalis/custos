<?php
namespace Acl {
    function in_array($needle, $haystack) { return false; }

    function check($who, array $roles) {
        if (<warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles)</warning>) { return 1; }
        return <warning descr="Use 'in_array(...)' to test membership.">array_search($who, $roles) !== false</warning>;
    }
}
