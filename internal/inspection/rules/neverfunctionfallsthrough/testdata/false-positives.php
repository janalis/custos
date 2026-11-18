<?php
function halt(): never { throw new RuntimeException(); }

// Loops whose condition is a constant expression never reach the closing brace.
function pollOnce(): never { do { usleep(1000); } while (1); }
function serve(): never { do { usleep(1000); } while (!false); }
function run(): never { while (!false) { usleep(1000); } }
function wait(): never { while (1 === 1) { usleep(1000); } }
function tick(int $i): never { for (; $i++, PHP_INT_SIZE > 2; ) { usleep(1000); } }
function spin(): never { while ((PHP_OS_FAMILY === 'Windows' ? 1 : 2) > 0) { usleep(1000); } }
function limits(): never { do { usleep(1000); } while (ArrayObject::ARRAY_AS_PROPS); }
