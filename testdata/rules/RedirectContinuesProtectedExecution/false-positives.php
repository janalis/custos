<?php
/** @custos-protected */ function purgeArchive() {}

function exclusiveBranch(bool $member) {
    if (!$member) { header('Location: /join'); }
    if ($member) { purgeArchive(); }
}

function laterGuard(bool $member) {
    if (!$member) { header('Location: /join'); }
    if (!$member) { return; }
    purgeArchive();
}

function guardedExpression(bool $member, bool $confirmed) {
    if (!$member) { header('Location: /join'); }
    $member && purgeArchive();
    !$member || purgeArchive();
    return $member && $confirmed ? purgeArchive() : null;
}

function otherwiseArm(bool $member) {
    if (!$member) { header('Location: /join'); }
    if (!$member) { echo 'redirecting'; } else { purgeArchive(); }
}

function reauthenticated(bool $member) {
    if (!$member) { header('Location: /join'); $member = (bool) getenv('FORCE_MEMBER'); }
    purgeArchive();
}

function deferredRedirect(bool $member) {
    if (!$member) { $later = function () { header('Location: /join'); }; }
    purgeArchive();
}

function deferredOperation(bool $member) {
    if (!$member) { header('Location: /join'); }
    return function () { purgeArchive(); };
}
