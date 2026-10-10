<?php
/** @custos-protected */ function deleteRecord() {}
function bad(bool $allowed) { if (!$allowed) { header('Location: /signin'); } <warning descr="Terminate the unauthorized branch after redirecting.">deleteRecord()</warning>; }
