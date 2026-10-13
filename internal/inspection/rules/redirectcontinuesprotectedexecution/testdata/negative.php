<?php
/** @custos-protected */ function deleteRecord() {}
function good(bool $allowed) { if (!$allowed) { header('Location: /signin'); return; } deleteRecord(); }
