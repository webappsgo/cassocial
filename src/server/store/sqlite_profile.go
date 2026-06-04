package store

import (
	"database/sql"
	"time"
)

// Profile operations
// Per PART 36: Profile is a user's public landing page

// GetProfileByID retrieves a profile by ID
func (db *DB) GetProfileByID(id string) (*Profile, error) {
	profile := &Profile{}
	err := db.QueryRowR(`
		SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
		       theme_id, custom_css, show_usernames, is_public, password_protected,
		       protection_password, custom_domain, domain_verified, analytics_enabled,
		       meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
		       created_at, updated_at
		FROM profiles WHERE id = ?
	`, id).Scan(
		&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
		&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
		&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
		&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfileBySlug retrieves a profile by slug
func (db *DB) GetProfileBySlug(slug string) (*Profile, error) {
	profile := &Profile{}
	err := db.QueryRowR(`
		SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
		       theme_id, custom_css, show_usernames, is_public, password_protected,
		       protection_password, custom_domain, domain_verified, analytics_enabled,
		       meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
		       created_at, updated_at
		FROM profiles WHERE slug = ?
	`, slug).Scan(
		&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
		&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
		&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
		&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfileByCustomDomain retrieves a profile by custom domain
func (db *DB) GetProfileByCustomDomain(domain string) (*Profile, error) {
	profile := &Profile{}
	err := db.QueryRowR(`
		SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
		       theme_id, custom_css, show_usernames, is_public, password_protected,
		       protection_password, custom_domain, domain_verified, analytics_enabled,
		       meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
		       created_at, updated_at
		FROM profiles WHERE custom_domain = ? AND domain_verified = 1
	`, domain).Scan(
		&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
		&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
		&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
		&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
		&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
		&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
		&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// GetProfilesByUserID retrieves all profiles for a user
func (db *DB) GetProfilesByUserID(userID string) ([]*Profile, error) {
	rows, err := db.QueryR(`
		SELECT id, user_id, slug, display_name, bio, avatar_url, header_image_url,
		       theme_id, custom_css, show_usernames, is_public, password_protected,
		       protection_password, custom_domain, domain_verified, analytics_enabled,
		       meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
		       created_at, updated_at
		FROM profiles WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*Profile
	for rows.Next() {
		profile := &Profile{}
		if err := rows.Scan(
			&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
			&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
			&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
			&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
			&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
			&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
			&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}

// CreateProfile creates a new profile
func (db *DB) CreateProfile(profile *Profile) error {
	_, err := db.ExecR(`
		INSERT INTO profiles (
			id, user_id, slug, display_name, bio, avatar_url, header_image_url,
			theme_id, custom_css, show_usernames, is_public, password_protected,
			protection_password, custom_domain, domain_verified, analytics_enabled,
			meta_title, meta_description, og_image_url, view_count, qr_code_enabled,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, profile.ID, profile.UserID, profile.Slug, profile.DisplayName, profile.Bio,
		profile.AvatarURL, profile.HeaderImageURL, profile.ThemeID, profile.CustomCSS,
		profile.ShowUsernames, profile.IsPublic, profile.PasswordProtected,
		profile.ProtectionPassword, profile.CustomDomain, profile.DomainVerified,
		profile.AnalyticsEnabled, profile.MetaTitle, profile.MetaDescription,
		profile.OgImageURL, profile.ViewCount, profile.QRCodeEnabled)
	return err
}

// UpdateProfile updates an existing profile
func (db *DB) UpdateProfile(profile *Profile) error {
	_, err := db.ExecR(`
		UPDATE profiles SET
			slug = ?, display_name = ?, bio = ?, avatar_url = ?, header_image_url = ?,
			theme_id = ?, custom_css = ?, show_usernames = ?, is_public = ?,
			password_protected = ?, protection_password = ?, custom_domain = ?,
			domain_verified = ?, analytics_enabled = ?, meta_title = ?,
			meta_description = ?, og_image_url = ?, qr_code_enabled = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, profile.Slug, profile.DisplayName, profile.Bio, profile.AvatarURL,
		profile.HeaderImageURL, profile.ThemeID, profile.CustomCSS, profile.ShowUsernames,
		profile.IsPublic, profile.PasswordProtected, profile.ProtectionPassword,
		profile.CustomDomain, profile.DomainVerified, profile.AnalyticsEnabled,
		profile.MetaTitle, profile.MetaDescription, profile.OgImageURL,
		profile.QRCodeEnabled, profile.ID)
	return err
}

// DeleteProfile deletes a profile
func (db *DB) DeleteProfile(id string) error {
	_, err := db.ExecR("DELETE FROM profiles WHERE id = ?", id)
	return err
}

// CountProfilesByUserID counts profiles for a user
func (db *DB) CountProfilesByUserID(userID string) (int, error) {
	var count int
	err := db.QueryRowR("SELECT COUNT(*) FROM profiles WHERE user_id = ?", userID).Scan(&count)
	return count, err
}

// IncrementProfileViewCount increments the view counter
func (db *DB) IncrementProfileViewCount(profileID string) error {
	_, err := db.ExecR("UPDATE profiles SET view_count = view_count + 1 WHERE id = ?", profileID)
	return err
}

// ProfileTheme operations

// GetProfileTheme retrieves theme settings for a profile
func (db *DB) GetProfileTheme(profileID string) (*ProfileTheme, error) {
	theme := &ProfileTheme{}
	err := db.QueryRowR(`
		SELECT profile_id, background_type, background_value, button_style,
		       button_animation, button_shadow, font_override, custom_css,
		       link_thumbnail_position, updated_at
		FROM profile_themes WHERE profile_id = ?
	`, profileID).Scan(
		&theme.ProfileID, &theme.BackgroundType, &theme.BackgroundValue,
		&theme.ButtonStyle, &theme.ButtonAnimation, &theme.ButtonShadow,
		&theme.FontOverride, &theme.CustomCSS, &theme.LinkThumbnailPosition,
		&theme.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return theme, nil
}

// UpdateProfileTheme updates or creates theme settings
func (db *DB) UpdateProfileTheme(theme *ProfileTheme) error {
	_, err := db.ExecR(`
		INSERT INTO profile_themes (
			profile_id, background_type, background_value, button_style,
			button_animation, button_shadow, font_override, custom_css,
			link_thumbnail_position, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id) DO UPDATE SET
			background_type = EXCLUDED.background_type,
			background_value = EXCLUDED.background_value,
			button_style = EXCLUDED.button_style,
			button_animation = EXCLUDED.button_animation,
			button_shadow = EXCLUDED.button_shadow,
			font_override = EXCLUDED.font_override,
			custom_css = EXCLUDED.custom_css,
			link_thumbnail_position = EXCLUDED.link_thumbnail_position,
			updated_at = CURRENT_TIMESTAMP
	`, theme.ProfileID, theme.BackgroundType, theme.BackgroundValue,
		theme.ButtonStyle, theme.ButtonAnimation, theme.ButtonShadow,
		theme.FontOverride, theme.CustomCSS, theme.LinkThumbnailPosition)
	return err
}

// DeleteProfileTheme deletes theme settings
func (db *DB) DeleteProfileTheme(profileID string) error {
	_, err := db.ExecR("DELETE FROM profile_themes WHERE profile_id = ?", profileID)
	return err
}

// QR Code Settings operations

// GetQRCodeSettings retrieves QR code settings
func (db *DB) GetQRCodeSettings(profileID string) (*QRCodeSettings, error) {
	settings := &QRCodeSettings{}
	err := db.QueryRowR(`
		SELECT profile_id, size, error_correction, style, dark_color, light_color,
		       logo_enabled, logo_size, format, updated_at
		FROM qr_code_settings WHERE profile_id = ?
	`, profileID).Scan(
		&settings.ProfileID, &settings.Size, &settings.ErrorCorrection,
		&settings.Style, &settings.DarkColor, &settings.LightColor,
		&settings.LogoEnabled, &settings.LogoSize, &settings.Format,
		&settings.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return settings, nil
}

// UpdateQRCodeSettings updates or creates QR code settings
func (db *DB) UpdateQRCodeSettings(settings *QRCodeSettings) error {
	_, err := db.ExecR(`
		INSERT INTO qr_code_settings (
			profile_id, size, error_correction, style, dark_color, light_color,
			logo_enabled, logo_size, format, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(profile_id) DO UPDATE SET
			size = EXCLUDED.size,
			error_correction = EXCLUDED.error_correction,
			style = EXCLUDED.style,
			dark_color = EXCLUDED.dark_color,
			light_color = EXCLUDED.light_color,
			logo_enabled = EXCLUDED.logo_enabled,
			logo_size = EXCLUDED.logo_size,
			format = EXCLUDED.format,
			updated_at = CURRENT_TIMESTAMP
	`, settings.ProfileID, settings.Size, settings.ErrorCorrection,
		settings.Style, settings.DarkColor, settings.LightColor,
		settings.LogoEnabled, settings.LogoSize, settings.Format)
	return err
}

// DeleteQRCodeSettings deletes QR code settings
func (db *DB) DeleteQRCodeSettings(profileID string) error {
	_, err := db.ExecR("DELETE FROM qr_code_settings WHERE profile_id = ?", profileID)
	return err
}

// Service operations
// Per PART 36: 5000+ predefined services

// GetServiceByID retrieves a service by ID
func (db *DB) GetServiceByID(id string) (*Service, error) {
	service := &Service{}
	err := db.QueryRowR(`
		SELECT id, name, category, icon_url, icon_svg, url_pattern,
		       background_color, text_color, popularity, is_active,
		       requires_username, placeholder_text, validation_pattern,
		       created_at, updated_at
		FROM services WHERE id = ?
	`, id).Scan(
		&service.ID, &service.Name, &service.Category, &service.IconURL,
		&service.IconSVG, &service.URLPattern, &service.BackgroundColor,
		&service.TextColor, &service.Popularity, &service.IsActive,
		&service.RequiresUsername, &service.PlaceholderText,
		&service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// GetServiceByName retrieves a service by name
func (db *DB) GetServiceByName(name string) (*Service, error) {
	service := &Service{}
	err := db.QueryRowR(`
		SELECT id, name, category, icon_url, icon_svg, url_pattern,
		       background_color, text_color, popularity, is_active,
		       requires_username, placeholder_text, validation_pattern,
		       created_at, updated_at
		FROM services WHERE name = ? AND is_active = 1
	`, name).Scan(
		&service.ID, &service.Name, &service.Category, &service.IconURL,
		&service.IconSVG, &service.URLPattern, &service.BackgroundColor,
		&service.TextColor, &service.Popularity, &service.IsActive,
		&service.RequiresUsername, &service.PlaceholderText,
		&service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// ListServices retrieves services with optional category filter
func (db *DB) ListServices(category string, limit, offset int) ([]*Service, error) {
	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = db.QueryR(`
			SELECT id, name, category, icon_url, icon_svg, url_pattern,
			       background_color, text_color, popularity, is_active,
			       requires_username, placeholder_text, validation_pattern,
			       created_at, updated_at
			FROM services
			WHERE category = ? AND is_active = 1
			ORDER BY popularity DESC, name ASC
			LIMIT ? OFFSET ?
		`, category, limit, offset)
	} else {
		rows, err = db.QueryR(`
			SELECT id, name, category, icon_url, icon_svg, url_pattern,
			       background_color, text_color, popularity, is_active,
			       requires_username, placeholder_text, validation_pattern,
			       created_at, updated_at
			FROM services
			WHERE is_active = 1
			ORDER BY popularity DESC, name ASC
			LIMIT ? OFFSET ?
		`, limit, offset)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*Service
	for rows.Next() {
		service := &Service{}
		if err := rows.Scan(
			&service.ID, &service.Name, &service.Category, &service.IconURL,
			&service.IconSVG, &service.URLPattern, &service.BackgroundColor,
			&service.TextColor, &service.Popularity, &service.IsActive,
			&service.RequiresUsername, &service.PlaceholderText,
			&service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt,
		); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, rows.Err()
}

// SearchServices searches services by name
func (db *DB) SearchServices(query string, limit int) ([]*Service, error) {
	rows, err := db.QueryR(`
		SELECT id, name, category, icon_url, icon_svg, url_pattern,
		       background_color, text_color, popularity, is_active,
		       requires_username, placeholder_text, validation_pattern,
		       created_at, updated_at
		FROM services
		WHERE is_active = 1 AND (name LIKE ? OR category LIKE ?)
		ORDER BY popularity DESC, name ASC
		LIMIT ?
	`, "%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []*Service
	for rows.Next() {
		service := &Service{}
		if err := rows.Scan(
			&service.ID, &service.Name, &service.Category, &service.IconURL,
			&service.IconSVG, &service.URLPattern, &service.BackgroundColor,
			&service.TextColor, &service.Popularity, &service.IsActive,
			&service.RequiresUsername, &service.PlaceholderText,
			&service.ValidationPattern, &service.CreatedAt, &service.UpdatedAt,
		); err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	return services, rows.Err()
}

// CreateService creates a new service
func (db *DB) CreateService(service *Service) error {
	_, err := db.ExecR(`
		INSERT INTO services (
			id, name, category, icon_url, icon_svg, url_pattern,
			background_color, text_color, popularity, is_active,
			requires_username, placeholder_text, validation_pattern,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, service.ID, service.Name, service.Category, service.IconURL,
		service.IconSVG, service.URLPattern, service.BackgroundColor,
		service.TextColor, service.Popularity, service.IsActive,
		service.RequiresUsername, service.PlaceholderText,
		service.ValidationPattern)
	return err
}

// UpdateService updates an existing service
func (db *DB) UpdateService(service *Service) error {
	_, err := db.ExecR(`
		UPDATE services SET
			name = ?, category = ?, icon_url = ?, icon_svg = ?, url_pattern = ?,
			background_color = ?, text_color = ?, popularity = ?, is_active = ?,
			requires_username = ?, placeholder_text = ?, validation_pattern = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, service.Name, service.Category, service.IconURL, service.IconSVG,
		service.URLPattern, service.BackgroundColor, service.TextColor,
		service.Popularity, service.IsActive, service.RequiresUsername,
		service.PlaceholderText, service.ValidationPattern, service.ID)
	return err
}

// DeleteService deletes a service
func (db *DB) DeleteService(id string) error {
	_, err := db.ExecR("DELETE FROM services WHERE id = ?", id)
	return err
}

// CountServices returns the total number of services
func (db *DB) CountServices() (int, error) {
	var count int
	err := db.QueryRowR("SELECT COUNT(*) FROM services WHERE is_active = 1").Scan(&count)
	return count, err
}

// Link operations
// Per PART 36: Links on user profiles

// GetLinkByID retrieves a link by ID
func (db *DB) GetLinkByID(id string) (*Link, error) {
	link := &Link{}
	err := db.QueryRowR(`
		SELECT id, profile_id, service_id, title, username, url, icon_url,
		       background_color, text_color, position, is_active, click_count,
		       created_at, updated_at
		FROM links WHERE id = ?
	`, id).Scan(
		&link.ID, &link.ProfileID, &link.ServiceID, &link.Title,
		&link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
		&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount,
		&link.CreatedAt, &link.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return link, nil
}

// GetLinksByProfileID retrieves all links for a profile
func (db *DB) GetLinksByProfileID(profileID string) ([]*Link, error) {
	rows, err := db.QueryR(`
		SELECT id, profile_id, service_id, title, username, url, icon_url,
		       background_color, text_color, position, is_active, click_count,
		       created_at, updated_at
		FROM links
		WHERE profile_id = ?
		ORDER BY position ASC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []*Link
	for rows.Next() {
		link := &Link{}
		if err := rows.Scan(
			&link.ID, &link.ProfileID, &link.ServiceID, &link.Title,
			&link.Username, &link.URL, &link.IconURL, &link.BackgroundColor,
			&link.TextColor, &link.Position, &link.IsActive, &link.ClickCount,
			&link.CreatedAt, &link.UpdatedAt,
		); err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	return links, rows.Err()
}

// CreateLink creates a new link
func (db *DB) CreateLink(link *Link) error {
	_, err := db.ExecR(`
		INSERT INTO links (
			id, profile_id, service_id, title, username, url, icon_url,
			background_color, text_color, position, is_active, click_count,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, link.ID, link.ProfileID, link.ServiceID, link.Title, link.Username,
		link.URL, link.IconURL, link.BackgroundColor, link.TextColor,
		link.Position, link.IsActive, link.ClickCount)
	return err
}

// UpdateLink updates an existing link
func (db *DB) UpdateLink(link *Link) error {
	_, err := db.ExecR(`
		UPDATE links SET
			service_id = ?, title = ?, username = ?, url = ?, icon_url = ?,
			background_color = ?, text_color = ?, position = ?, is_active = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, link.ServiceID, link.Title, link.Username, link.URL, link.IconURL,
		link.BackgroundColor, link.TextColor, link.Position, link.IsActive, link.ID)
	return err
}

// DeleteLink deletes a link
func (db *DB) DeleteLink(id string) error {
	_, err := db.ExecR("DELETE FROM links WHERE id = ?", id)
	return err
}

// ReorderLinks updates link positions
func (db *DB) ReorderLinks(profileID string, linkIDs []string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for position, linkID := range linkIDs {
		_, err := tx.Exec("UPDATE links SET position = ? WHERE id = ? AND profile_id = ?",
			position, linkID, profileID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// CountLinksByProfileID counts links for a profile
func (db *DB) CountLinksByProfileID(profileID string) (int, error) {
	var count int
	err := db.QueryRowR("SELECT COUNT(*) FROM links WHERE profile_id = ?", profileID).Scan(&count)
	return count, err
}

// IncrementLinkClickCount increments the click counter
func (db *DB) IncrementLinkClickCount(linkID string) error {
	_, err := db.ExecR("UPDATE links SET click_count = click_count + 1 WHERE id = ?", linkID)
	return err
}

// FooterItem operations

// GetFooterItemsByProfileID retrieves footer items for a profile
func (db *DB) GetFooterItemsByProfileID(profileID string) ([]*FooterItem, error) {
	rows, err := db.QueryR(`
		SELECT id, profile_id, item_type, content, position, is_active, created_at
		FROM footer_items
		WHERE profile_id = ?
		ORDER BY position ASC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*FooterItem
	for rows.Next() {
		item := &FooterItem{}
		if err := rows.Scan(
			&item.ID, &item.ProfileID, &item.ItemType, &item.Content,
			&item.Position, &item.IsActive, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// CreateFooterItem creates a new footer item
func (db *DB) CreateFooterItem(item *FooterItem) error {
	_, err := db.ExecR(`
		INSERT INTO footer_items (id, profile_id, item_type, content, position, is_active, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, item.ID, item.ProfileID, item.ItemType, item.Content, item.Position, item.IsActive)
	return err
}

// UpdateFooterItem updates an existing footer item
func (db *DB) UpdateFooterItem(item *FooterItem) error {
	_, err := db.ExecR(`
		UPDATE footer_items SET
			item_type = ?, content = ?, position = ?, is_active = ?
		WHERE id = ?
	`, item.ItemType, item.Content, item.Position, item.IsActive, item.ID)
	return err
}

// DeleteFooterItem deletes a footer item
func (db *DB) DeleteFooterItem(id string) error {
	_, err := db.ExecR("DELETE FROM footer_items WHERE id = ?", id)
	return err
}

// Shortlink operations
// Per PART 36: URL shortener functionality

// GetShortlinkByID retrieves a shortlink by ID
func (db *DB) GetShortlinkByID(id string) (*Shortlink, error) {
	shortlink := &Shortlink{}
	err := db.QueryRowR(`
		SELECT id, short_code, target_url, profile_id, title, click_count, expires_at, created_at
		FROM shortlinks WHERE id = ?
	`, id).Scan(
		&shortlink.ID, &shortlink.ShortCode, &shortlink.TargetURL,
		&shortlink.ProfileID, &shortlink.Title, &shortlink.ClickCount,
		&shortlink.ExpiresAt, &shortlink.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return shortlink, nil
}

// GetShortlinkByCode retrieves a shortlink by short code
func (db *DB) GetShortlinkByCode(code string) (*Shortlink, error) {
	shortlink := &Shortlink{}
	err := db.QueryRowR(`
		SELECT id, short_code, target_url, profile_id, title, click_count, expires_at, created_at
		FROM shortlinks WHERE short_code = ?
	`, code).Scan(
		&shortlink.ID, &shortlink.ShortCode, &shortlink.TargetURL,
		&shortlink.ProfileID, &shortlink.Title, &shortlink.ClickCount,
		&shortlink.ExpiresAt, &shortlink.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return shortlink, nil
}

// GetShortlinksByProfileID retrieves all shortlinks for a profile
func (db *DB) GetShortlinksByProfileID(profileID string) ([]*Shortlink, error) {
	rows, err := db.QueryR(`
		SELECT id, short_code, target_url, profile_id, title, click_count, expires_at, created_at
		FROM shortlinks
		WHERE profile_id = ?
		ORDER BY created_at DESC
	`, profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var shortlinks []*Shortlink
	for rows.Next() {
		shortlink := &Shortlink{}
		if err := rows.Scan(
			&shortlink.ID, &shortlink.ShortCode, &shortlink.TargetURL,
			&shortlink.ProfileID, &shortlink.Title, &shortlink.ClickCount,
			&shortlink.ExpiresAt, &shortlink.CreatedAt,
		); err != nil {
			return nil, err
		}
		shortlinks = append(shortlinks, shortlink)
	}

	return shortlinks, rows.Err()
}

// CreateShortlink creates a new shortlink
func (db *DB) CreateShortlink(shortlink *Shortlink) error {
	_, err := db.ExecR(`
		INSERT INTO shortlinks (id, short_code, target_url, profile_id, title, click_count, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, shortlink.ID, shortlink.ShortCode, shortlink.TargetURL, shortlink.ProfileID,
		shortlink.Title, shortlink.ClickCount, db.BindNullableTime(shortlink.ExpiresAt))
	return err
}

// UpdateShortlink updates an existing shortlink
func (db *DB) UpdateShortlink(shortlink *Shortlink) error {
	_, err := db.ExecR(`
		UPDATE shortlinks SET
			target_url = ?, title = ?, expires_at = ?
		WHERE id = ?
	`, shortlink.TargetURL, shortlink.Title, db.BindNullableTime(shortlink.ExpiresAt), shortlink.ID)
	return err
}

// DeleteShortlink deletes a shortlink
func (db *DB) DeleteShortlink(id string) error {
	_, err := db.ExecR("DELETE FROM shortlinks WHERE id = ?", id)
	return err
}

// IncrementShortlinkClickCount increments the click counter
func (db *DB) IncrementShortlinkClickCount(id string) error {
	_, err := db.ExecR("UPDATE shortlinks SET click_count = click_count + 1 WHERE id = ?", id)
	return err
}

// DeleteExpiredShortlinks removes expired shortlinks
func (db *DB) DeleteExpiredShortlinks() error {
	_, err := db.ExecR("DELETE FROM shortlinks WHERE expires_at IS NOT NULL AND expires_at < datetime('now')")
	return err
}

// Analytics operations
// Per PART 36: Analytics tracking with hashed IPs for GDPR

// RecordProfileView records a profile view
func (db *DB) RecordProfileView(view *ProfileView) error {
	_, err := db.ExecR(`
		INSERT INTO profile_views (profile_id, viewer_ip, referrer, user_agent, country, timestamp)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, view.ProfileID, view.ViewerIP, view.Referrer, view.UserAgent, view.Country)
	return err
}

// RecordLinkClick records a link click
func (db *DB) RecordLinkClick(click *LinkClick) error {
	_, err := db.ExecR(`
		INSERT INTO link_clicks (link_id, clicker_ip, referrer, user_agent, country, timestamp)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, click.LinkID, click.ClickerIP, click.Referrer, click.UserAgent, click.Country)
	return err
}

// GetProfileAnalytics retrieves analytics for a profile
func (db *DB) GetProfileAnalytics(profileID string, days int) (*ProfileAnalytics, error) {
	analytics := &ProfileAnalytics{ProfileID: profileID}

	cutoffDate := time.Now().AddDate(0, 0, -days)

	err := db.QueryRowR(`
		SELECT COUNT(*), COUNT(DISTINCT viewer_ip)
		FROM profile_views
		WHERE profile_id = ? AND timestamp >= ?
	`, profileID, cutoffDate).Scan(&analytics.Views, &analytics.UniqueIPs)
	if err != nil {
		return nil, err
	}

	err = db.QueryRowR(`
		SELECT COUNT(*)
		FROM link_clicks
		WHERE link_id IN (SELECT id FROM links WHERE profile_id = ?)
		AND timestamp >= ?
	`, profileID, cutoffDate).Scan(&analytics.Clicks)
	if err != nil {
		return nil, err
	}

	topLinks, err := db.GetTopLinks(profileID, 10)
	if err == nil {
		analytics.TopLinks = topLinks
	}

	topReferrers, err := db.GetTopReferrers(profileID, 10)
	if err == nil {
		analytics.TopReferrers = topReferrers
	}

	return analytics, nil
}

// GetLinkAnalytics retrieves analytics for a link
func (db *DB) GetLinkAnalytics(linkID string, days int) (*LinkAnalytics, error) {
	analytics := &LinkAnalytics{LinkID: linkID}

	cutoffDate := time.Now().AddDate(0, 0, -days)

	err := db.QueryRowR(`
		SELECT COUNT(*), COUNT(DISTINCT clicker_ip)
		FROM link_clicks
		WHERE link_id = ? AND timestamp >= ?
	`, linkID, cutoffDate).Scan(&analytics.Clicks, &analytics.UniqueIPs)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryR(`
		SELECT referrer, COUNT(*) as count
		FROM link_clicks
		WHERE link_id = ? AND timestamp >= ? AND referrer != ''
		GROUP BY referrer
		ORDER BY count DESC
		LIMIT 10
	`, linkID, cutoffDate)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat ReferrerStat
			if err := rows.Scan(&stat.Referrer, &stat.Count); err == nil {
				analytics.TopReferrers = append(analytics.TopReferrers, &stat)
			}
		}
	}

	return analytics, nil
}

// GetTopLinks retrieves top links by clicks
func (db *DB) GetTopLinks(profileID string, limit int) ([]*LinkStat, error) {
	rows, err := db.QueryR(`
		SELECT l.id, l.title, COUNT(lc.id) as clicks
		FROM links l
		LEFT JOIN link_clicks lc ON l.id = lc.link_id
		WHERE l.profile_id = ?
		GROUP BY l.id, l.title
		ORDER BY clicks DESC
		LIMIT ?
	`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*LinkStat
	for rows.Next() {
		stat := &LinkStat{}
		if err := rows.Scan(&stat.LinkID, &stat.Title, &stat.Clicks); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// GetTopReferrers retrieves top referrers
func (db *DB) GetTopReferrers(profileID string, limit int) ([]*ReferrerStat, error) {
	rows, err := db.QueryR(`
		SELECT pv.referrer, COUNT(*) as count
		FROM profile_views pv
		WHERE pv.profile_id = ? AND pv.referrer != ''
		GROUP BY pv.referrer
		ORDER BY count DESC
		LIMIT ?
	`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []*ReferrerStat
	for rows.Next() {
		stat := &ReferrerStat{}
		if err := rows.Scan(&stat.Referrer, &stat.Count); err != nil {
			return nil, err
		}
		stats = append(stats, stat)
	}

	return stats, rows.Err()
}

// Cluster operations
// Per PART 24: Cluster support with heartbeat monitoring

// CreateClusterNode creates a new cluster node
func (db *DB) CreateClusterNode(node *ClusterNode) error {
	_, err := db.ExecR(`
		INSERT INTO cluster_nodes (id, hostname, address, port, status, is_primary, last_heartbeat, created_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`, node.ID, node.Hostname, node.Address, node.Port, node.Status, node.IsPrimary)
	return err
}

// UpdateClusterNode updates a cluster node
func (db *DB) UpdateClusterNode(node *ClusterNode) error {
	_, err := db.ExecR(`
		UPDATE cluster_nodes SET
			hostname = ?, address = ?, port = ?, status = ?, is_primary = ?, last_heartbeat = ?
		WHERE id = ?
	`, node.Hostname, node.Address, node.Port, node.Status, node.IsPrimary, node.LastHeartbeat, node.ID)
	return err
}

// GetClusterNode retrieves a cluster node by ID
func (db *DB) GetClusterNode(id string) (*ClusterNode, error) {
	node := &ClusterNode{}
	err := db.QueryRowR(`
		SELECT id, hostname, address, port, status, is_primary, last_heartbeat, created_at
		FROM cluster_nodes WHERE id = ?
	`, id).Scan(
		&node.ID, &node.Hostname, &node.Address, &node.Port,
		&node.Status, &node.IsPrimary, &node.LastHeartbeat, &node.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// ListClusterNodes retrieves all cluster nodes
func (db *DB) ListClusterNodes() ([]*ClusterNode, error) {
	rows, err := db.QueryR(`
		SELECT id, hostname, address, port, status, is_primary, last_heartbeat, created_at
		FROM cluster_nodes
		ORDER BY is_primary DESC, created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nodes []*ClusterNode
	for rows.Next() {
		node := &ClusterNode{}
		if err := rows.Scan(
			&node.ID, &node.Hostname, &node.Address, &node.Port,
			&node.Status, &node.IsPrimary, &node.LastHeartbeat, &node.CreatedAt,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}

	return nodes, rows.Err()
}

// UpdateNodeHeartbeat updates a node's heartbeat timestamp
func (db *DB) UpdateNodeHeartbeat(id string) error {
	_, err := db.ExecR("UPDATE cluster_nodes SET last_heartbeat = CURRENT_TIMESTAMP, status = 'healthy' WHERE id = ?", id)
	return err
}

// DeleteClusterNode deletes a cluster node
func (db *DB) DeleteClusterNode(id string) error {
	_, err := db.ExecR("DELETE FROM cluster_nodes WHERE id = ?", id)
	return err
}

// GetPrimaryNode retrieves the primary cluster node
func (db *DB) GetPrimaryNode() (*ClusterNode, error) {
	node := &ClusterNode{}
	err := db.QueryRowR(`
		SELECT id, hostname, address, port, status, is_primary, last_heartbeat, created_at
		FROM cluster_nodes WHERE is_primary = 1
	`).Scan(
		&node.ID, &node.Hostname, &node.Address, &node.Port,
		&node.Status, &node.IsPrimary, &node.LastHeartbeat, &node.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return node, nil
}

// MarkNodeOffline marks a node as offline
func (db *DB) MarkNodeOffline(id string) error {
	_, err := db.ExecR("UPDATE cluster_nodes SET status = 'offline' WHERE id = ?", id)
	return err
}

// Profile Tags operations

// AddProfileTag adds a tag to a profile
func (db *DB) AddProfileTag(profileID, tag string) error {
	_, err := db.ExecR("INSERT OR IGNORE INTO profile_tags (profile_id, tag) VALUES (?, ?)", profileID, tag)
	return err
}

// RemoveProfileTag removes a tag from a profile
func (db *DB) RemoveProfileTag(profileID, tag string) error {
	_, err := db.ExecR("DELETE FROM profile_tags WHERE profile_id = ? AND tag = ?", profileID, tag)
	return err
}

// GetProfileTags retrieves all tags for a profile
func (db *DB) GetProfileTags(profileID string) ([]string, error) {
	rows, err := db.QueryR("SELECT tag FROM profile_tags WHERE profile_id = ? ORDER BY tag", profileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// SearchProfilesByTag searches profiles by tag
func (db *DB) SearchProfilesByTag(tag string, limit, offset int) ([]*Profile, error) {
	rows, err := db.QueryR(`
		SELECT p.id, p.user_id, p.slug, p.display_name, p.bio, p.avatar_url, p.header_image_url,
		       p.theme_id, p.custom_css, p.show_usernames, p.is_public, p.password_protected,
		       p.protection_password, p.custom_domain, p.domain_verified, p.analytics_enabled,
		       p.meta_title, p.meta_description, p.og_image_url, p.view_count, p.qr_code_enabled,
		       p.created_at, p.updated_at
		FROM profiles p
		INNER JOIN profile_tags pt ON p.id = pt.profile_id
		WHERE pt.tag = ? AND p.is_public = 1
		ORDER BY p.created_at DESC
		LIMIT ? OFFSET ?
	`, tag, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var profiles []*Profile
	for rows.Next() {
		profile := &Profile{}
		if err := rows.Scan(
			&profile.ID, &profile.UserID, &profile.Slug, &profile.DisplayName,
			&profile.Bio, &profile.AvatarURL, &profile.HeaderImageURL, &profile.ThemeID,
			&profile.CustomCSS, &profile.ShowUsernames, &profile.IsPublic,
			&profile.PasswordProtected, &profile.ProtectionPassword, &profile.CustomDomain,
			&profile.DomainVerified, &profile.AnalyticsEnabled, &profile.MetaTitle,
			&profile.MetaDescription, &profile.OgImageURL, &profile.ViewCount,
			&profile.QRCodeEnabled, &profile.CreatedAt, &profile.UpdatedAt,
		); err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, rows.Err()
}
