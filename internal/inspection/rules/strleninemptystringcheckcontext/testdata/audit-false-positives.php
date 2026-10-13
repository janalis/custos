<?php
function audit_case_0() { mb_strlen('a', encoding()) > 0; }
function audit_case_1() { mb_strlen("\x1b(B", 'ISO-2022-JP') === 0; }
function audit_case_2() { if (mb_strlen($s)) {} }
function audit_case_3() { strlen('a', extra()) > 0; }
