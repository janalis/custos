<?php
/** @custos-protected */ function deleteRecord() {}
function bad(bool $allowed) { if (!$allowed) { header('Location: /signin'); } <warning descr="Terminate the unauthorized branch after redirecting.">deleteRecord()</warning>; }
function unrelatedGuard(bool $allowed, bool $confirmed) {
    if (!$allowed) { header('Location: /signin'); }
    if (!$confirmed) { return; }
    if ($confirmed) { <warning descr="Terminate the unauthorized branch after redirecting.">deleteRecord()</warning>; }
}
if (!$isAdmin) { header('Location: /signin'); }
<warning descr="Terminate the unauthorized branch after redirecting.">deleteRecord()</warning>;
