<?php
// Strings PHP treats as numeric keep loose semantics: no strict fix offered.
function numbers($v) {
    return [
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == '1e3'</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == ' 1'</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == '1.'</weak_warning>,
        <weak_warning descr="Prefer '!==' to avoid implicit type juggling.">$v != '-2.5E-3'</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == "7\n"</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == "\x31"</weak_warning>,
        <weak_warning descr="Prefer '===' to avoid implicit type juggling.">$v == "{$v}"</weak_warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$v == '1e'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$v == '.'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$v == '0x1A'</warning>,
        <warning descr="Use '===' here; the string is not numeric, so strict comparison is safe.">$v == '1_000'</warning>,
    ];
}
