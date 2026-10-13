<?php
// Guard clauses, allow-lists and compound conditions that accept an upload on its client label alone.
function guardReturn() {
    if ($_FILES['f']['type'] !== 'image/png') { return; }
    <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
}
function guardThrowLoose() {
    if ('image/png' != $_FILES['f']['type']) { throw new RuntimeException('type'); }
    if ($_FILES['f']['error'] !== UPLOAD_ERR_OK) { return; }
    foreach ([1, 2] as $attempt) {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function guardDisjunction() {
    if ($_FILES['f']['size'] > 1000 || $_FILES['f']['type'] !== 'image/png') { exit; }
    <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
}
function guardNegatedAllowList() {
    if (!in_array($_FILES['f']['type'], ['image/png', 'image/jpeg'], true)) { return; }
    <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
}
function guardNegatedConjunction() {
    if (!($_FILES['f']['type'] === 'image/png' && $_FILES['f']['size'] < 1000)) { return; }
    <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
}
function allowList() {
    if (in_array($_FILES['f']['type'], ['image/png', 'image/jpeg'], true)) {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function allowListArraySyntax() {
    if (in_array(haystack: array('image/png'), needle: $_FILES['f']['type'])) {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function conjunction() {
    if ($_FILES['f']['size'] < 1000 && ($_FILES['f']['type'] == 'image/png')) {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function negatedDisjunction() {
    if (!($_FILES['f']['type'] !== 'image/png' || $_FILES['f']['size'] > 1000)) {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function elseIfBranch() {
    if ($_FILES['f']['error'] !== UPLOAD_ERR_OK) {
        return;
    } elseif ($_FILES['f']['type'] === 'image/png') {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function unbraced() {
    if ($_FILES['f']['type'] === 'image/png')
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
}
function errorCheckedOnly() {
    if ($_FILES['f']['type'] === 'image/png') {
        if ($_FILES['f']['error'] !== UPLOAD_ERR_OK) { return; }
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function otherUploadInspected() {
    $other = $_FILES['g']['tmp_name'];
    if (getimagesize($other) === false || mime_content_type($_FILES['g']['tmp_name']) !== 'image/png') { return; }
    if ($_FILES['f']['type'] === 'image/png') {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function inspectedTooLate() {
    if ($_FILES['f']['type'] === 'image/png') {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
        getimagesize($_FILES['f']['tmp_name']);
    }
}
function unrelatedCalls($storage) {
    $storage->{'file'}($_FILES['f']['tmp_name']);
    $storage->path($_FILES['f']['tmp_name']);
    is_uploaded_file($_FILES['f']['tmp_name']);
    if ($_FILES['f']['type'] === 'image/png') {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function probeNeverApplied() {
    $probe = getimagesize(...);
    if ($_FILES['f']['type'] === 'image/png') {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function elseAfterMismatch() {
    if ($_FILES['f']['type'] !== 'image/png') {
        error_log('unexpected type');
    } else {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function elseIfAfterMismatch($ready) {
    if (!in_array($_FILES['f']['type'], ['image/png', 'image/gif'])) {
        error_log('unexpected type');
    } elseif ($ready) {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function elseAfterMismatchingElseIf($ready) {
    if (!$ready) {
        return;
    } elseif ($_FILES['f']['type'] != 'image/png') {
        error_log('unexpected type');
    } else if ($_FILES['f']['size'] < 1000) {
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    }
}
function alternativeSyntax() {
    if ($_FILES['f']['type'] !== 'image/png'):
        error_log('unexpected type');
    else:
        <warning descr="Inspect uploaded content instead of trusting its client MIME label.">move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png')</warning>;
    endif;
}
