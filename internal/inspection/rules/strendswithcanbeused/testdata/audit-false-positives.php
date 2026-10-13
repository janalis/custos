<?php
function audit_case_0() { mb_substr('aé', -strlen('é')) === 'é'; }
function audit_case_1() { substr('aé', -mb_strlen('é')) === 'é'; }
function audit_case_2() { mb_substr('aé', -mb_strlen('é')) === 'é'; }
