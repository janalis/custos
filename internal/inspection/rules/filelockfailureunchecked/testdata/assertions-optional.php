<?php
function writeAfterOptionalLock($handle) {
    assert(flock($handle, LOCK_EX));
    <warning descr="Acquire the lock successfully before writing.">fwrite($handle, 'payload')</warning>;
}
