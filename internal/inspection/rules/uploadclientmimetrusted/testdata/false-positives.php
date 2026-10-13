<?php
// Server-side content inspection of the same temporary file accompanies the client label.
function nestedFinfo() {
    if ($_FILES['f']['type'] === 'image/png') {
        if ((new finfo(FILEINFO_MIME_TYPE))->file($_FILES['f']['tmp_name']) === 'image/png') {
            move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png');
        }
    }
}
function conjoinedFinfo() {
    if ($_FILES['f']['type'] === 'image/png' && (new finfo(FILEINFO_MIME_TYPE))->file($_FILES['f']['tmp_name']) === 'image/png') {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png');
    }
}
function imageSizeGuard() {
    if ($_FILES['f']['type'] === 'image/png') {
        if (getimagesize($_FILES['f']['tmp_name']) === false) { return; }
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png');
    }
}
function earlierMimeGuard() {
    if (mime_content_type($_FILES['f']['tmp_name']) !== 'image/png') { return; }
    if ($_FILES['f']['type'] !== 'image/png') { return; }
    move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png');
}
function decoded() {
    if ($_FILES['f']['type'] === 'image/png') {
        $image = imagecreatefrompng($_FILES['f']['tmp_name']);
        if (!$image) { return; }
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png');
    }
}
function exifGuard() {
    if (exif_imagetype($_FILES['f']['tmp_name']) !== IMAGETYPE_PNG) { throw new RuntimeException('type'); }
    if (in_array($_FILES['f']['type'], ['image/png'], true)) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png');
    }
}
function aliasedTemporaryName() {
    $tmp = $_FILES['f']['tmp_name'];
    $info = finfo_open(FILEINFO_MIME_TYPE);
    if (finfo_file($info, $tmp) !== 'image/png') { return; }
    if ($_FILES['f']['type'] === 'image/png') {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/image.png');
    }
}

// The client label does not decide acceptance of this upload.
function rejectingGuard() {
    if ($_FILES['f']['type'] === 'image/png') { return; }
    move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
}
function nonTerminatingGuard() {
    if ($_FILES['f']['type'] !== 'image/png') { error_log('unexpected type'); }
    move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
}
function rejectedBranch() {
    if ($_FILES['f']['type'] !== 'image/png') {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/quarantine/a.bin');
    }
}
function alternativeAcceptance() {
    if ($_FILES['f']['type'] === 'image/png' || $_FILES['f']['size'] < 10) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
}
function guardNeedsBoth() {
    if ($_FILES['f']['type'] !== 'image/png' && $_FILES['f']['size'] > 10) { return; }
    move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
}
function extensionOnly() {
    if ($_FILES['f']['type'] === 'image/png') { $ext = 'png'; } else { $ext = 'bin'; }
    move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.' . $ext);
}
function notMimeValues() {
    if ($_FILES['f']['type'] === 'png') {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
    if (in_array($_FILES['f']['type'], ['image/png', $extra])) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
    if (in_array($_FILES['f']['type'], [])) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
    if (in_array($_FILES['f']['name'], ['image/png'])) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
}
function otherConditions($ready) {
    if (-$ready) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
    if ($_FILES['f']['size'] < 10) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
    if ($ready) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    } elseif ($_FILES['f']['size'] < 10) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
    if (is_string($_FILES['f']['type'])) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
}
function otherSources($other) {
    if ($_FILES['f']['type'] === 'image/png') {
        move_uploaded_file($_FILES['tmp_name'], '/srv/uploads/a.bin');
        move_uploaded_file($other['f']['tmp_name'], '/srv/uploads/a.bin');
        move_uploaded_file($_FILES['f']['name'], '/srv/uploads/a.bin');
        move_uploaded_file($other, '/srv/uploads/a.bin');
        rename($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    }
}
function alternativesWithoutMismatch($ready, $later) {
    if ($ready) {
        return;
    } elseif ($later) {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/uploads/a.bin');
    } elseif ($_FILES['f']['type'] !== 'image/png') {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/quarantine/a.bin');
    } else {
        error_log('nothing stored');
    }
    if ($_FILES['f']['type'] === 'image/png') {
        error_log('known type');
    } else {
        move_uploaded_file($_FILES['f']['tmp_name'], '/srv/quarantine/a.bin');
    }
    if ($_FILES['f']['type'] !== 'image/png' && move_uploaded_file($_FILES['f']['tmp_name'], '/srv/quarantine/a.bin')) {
        return;
    }
}
