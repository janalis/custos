<?php
namespace Acl {
    function in_array($needle, $haystack) { return false; }

    function check($who, array $roles) {
        if (\in_array($who, $roles)) { return 1; }
        return \in_array($who, $roles);
    }
}
