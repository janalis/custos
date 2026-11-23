<?php
function find($ldap) {
    // A partial application passes the arguments it binds on every call.
    $search = ldap_search($ldap, 'dc=x', <warning descr="Escape request data for LDAP filter syntax.">'(uid=' . $_GET['u'] . ')'</warning>, ...);
    // The filter comes with each call.
    $later = ldap_search($ldap, 'dc=x', ?);
}
