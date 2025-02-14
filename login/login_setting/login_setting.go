package login_setting

import (
	"errors"
	"fmt"
	"r3/db"
	"r3/types"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func Get(loginId pgtype.Int8, loginTemplateId pgtype.Int8) (types.Settings, error) {

	var s types.Settings
	if (loginId.Valid && loginTemplateId.Valid) || (!loginId.Valid && !loginTemplateId.Valid) {
		return s, errors.New("settings can only be retrieved for either login or login template")
	}

	entryId := loginId.Int64
	entryName := "login_id"

	if loginTemplateId.Valid {
		entryId = loginTemplateId.Int64
		entryName = "login_template_id"
	}

	err := db.Pool.QueryRow(db.Ctx, fmt.Sprintf(`
		SELECT language_code, date_format, sunday_first_dow, font_size, borders_all,
			borders_corner, page_limit, header_captions, spacing, dark, compact,
			hint_update_version, mobile_scroll_form, warn_unsaved, menu_colored,
			pattern, font_family, tab_remember, field_clean, ARRAY(
				SELECT name::TEXT
				FROM instance.login_search_dict
				WHERE login_id          = ls.login_id
				OR    login_template_id = ls.login_template_id
				ORDER BY position ASC
			)
		FROM instance.login_setting AS ls
		WHERE %s = $1
	`, entryName), entryId).Scan(&s.LanguageCode, &s.DateFormat, &s.SundayFirstDow,
		&s.FontSize, &s.BordersAll, &s.BordersCorner, &s.PageLimit,
		&s.HeaderCaptions, &s.Spacing, &s.Dark, &s.Compact, &s.HintUpdateVersion,
		&s.MobileScrollForm, &s.WarnUnsaved, &s.MenuColored, &s.Pattern,
		&s.FontFamily, &s.TabRemember, &s.FieldClean, &s.SearchDictionaries)

	return s, err
}

func Set_tx(tx pgx.Tx, loginId pgtype.Int8, loginTemplateId pgtype.Int8, s types.Settings, isNew bool) error {

	if (loginId.Valid && loginTemplateId.Valid) || (!loginId.Valid && !loginTemplateId.Valid) {
		return errors.New("settings can only be applied for either login or login template")
	}

	entryId := loginId.Int64
	entryName := "login_id"

	if loginTemplateId.Valid {
		entryId = loginTemplateId.Int64
		entryName = "login_template_id"
	}

	if isNew {
		if _, err := tx.Exec(db.Ctx, fmt.Sprintf(`
			INSERT INTO instance.login_setting (%s, language_code, date_format,
				sunday_first_dow, font_size, borders_all, borders_corner, page_limit, 
				header_captions, spacing, dark, compact, hint_update_version,
				mobile_scroll_form, warn_unsaved, menu_colored, pattern, font_family,
				tab_remember, field_clean)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		`, entryName), entryId, s.LanguageCode, s.DateFormat, s.SundayFirstDow,
			s.FontSize, s.BordersAll, s.BordersCorner, s.PageLimit,
			s.HeaderCaptions, s.Spacing, s.Dark, s.Compact, s.HintUpdateVersion,
			s.MobileScrollForm, s.WarnUnsaved, s.MenuColored, s.Pattern,
			s.FontFamily, s.TabRemember, s.FieldClean); err != nil {

			return err
		}
	} else {
		if _, err := tx.Exec(db.Ctx, fmt.Sprintf(`
			UPDATE instance.login_setting
			SET language_code = $1, date_format = $2, sunday_first_dow = $3,
				font_size = $4, borders_all = $5, borders_corner = $6,
				page_limit = $7, header_captions = $8, spacing = $9, dark = $10,
				compact = $11, hint_update_version = $12, mobile_scroll_form = $13,
				warn_unsaved = $14, menu_colored = $15, pattern = $16,
				font_family = $17, tab_remember = $18, field_clean = $19
			WHERE %s = $20
		`, entryName), s.LanguageCode, s.DateFormat, s.SundayFirstDow, s.FontSize, s.BordersAll,
			s.BordersCorner, s.PageLimit, s.HeaderCaptions, s.Spacing, s.Dark,
			s.Compact, s.HintUpdateVersion, s.MobileScrollForm, s.WarnUnsaved,
			s.MenuColored, s.Pattern, s.FontFamily, s.TabRemember, s.FieldClean,
			entryId); err != nil {

			return err
		}
	}

	// update full text search dictionaries
	if !isNew {
		if _, err := tx.Exec(db.Ctx, fmt.Sprintf(`
			DELETE FROM instance.login_search_dict
			WHERE %s = $1
		`, entryName), entryId); err != nil {
			return err
		}
	}

	for i, dictName := range s.SearchDictionaries {
		if _, err := tx.Exec(db.Ctx, fmt.Sprintf(`
			INSERT INTO instance.login_search_dict (%s, position, name)
			VALUES ($1, $2, $3)
		`, entryName), entryId, i, dictName); err != nil {
			return err
		}
	}
	return nil
}
