<?php
function audit() {
    /** <weak_warning descr="Suppressed inspection not recognized: PhpLegacyCheckInspector.">@noinspection</weak_warning> PhpLegacyCheckInspector */
    /* <weak_warning descr="Suppressed inspection not recognized: TypoNameInspection.">@noinspection</weak_warning> TypoNameInspection */
    /** <weak_warning descr="Suppressed inspection not recognized: PhpGoneAway.">@noinspection</weak_warning> PhpUnused, PhpGoneAway -- kept for BC */
    /**
     * Two tags in one comment.
     * <weak_warning descr="Suppressed inspection not recognized: PhpMissing, OtherInspection.">@noinspection</weak_warning> PhpMissing OtherInspection PhpMissing
     * @noinspection PhpUnusedLocalVariableInspection
     * <weak_warning descr="Suppressed inspection not recognized: UnsetMergedInspection.">@noinspection</weak_warning> UnsetMergedInspection */
    return 1;
}
