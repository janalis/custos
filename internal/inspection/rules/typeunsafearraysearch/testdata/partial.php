<?php
// A partial application binds the missing strict flag for every call.
$find = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">in_array(?, $haystack)</weak_warning>;
$where = <weak_warning descr="Pass a third argument to say whether this search must be type-strict.">array_search($needle, ?)</weak_warning>;
$strict = in_array(?, $haystack, true);
// The strict flag comes with each call.
$later = in_array(?, $haystack, ...);
$named = in_array(?, haystack: ?);
$closure = in_array(...);
// A haystack of non-numeric strings compares like a strict search.
$letters = in_array(?, ['a', 'b']);
// A named argument leaves the binding to PHP's rules for names.
$byName = in_array(?, haystack: $haystack);
