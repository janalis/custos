<?php
// @custos-ignore AsyncSignalSetterReturnMisread
if (!pcntl_async_signals(true)) { throw new RuntimeException(); }
