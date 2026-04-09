-- Migration 014: Legal Pages (Terms of Service, Privacy Policy, etc.)

CREATE TABLE IF NOT EXISTS legal_pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_legal_pages_slug ON legal_pages(slug);
CREATE INDEX IF NOT EXISTS idx_legal_pages_is_active ON legal_pages(is_active);

-- Seed default pages
INSERT INTO legal_pages (slug, title, content) VALUES
    ('terms', 'Terms of Service', '<h2>Terms of Service</h2>
<p>Last updated: [DATE]</p>

<h3>1. Acceptance of Terms</h3>
<p>By accessing and using this service, you accept and agree to be bound by the terms and provisions of this agreement.</p>

<h3>2. Use of Service</h3>
<p>You agree to use the service only for lawful purposes and in accordance with these Terms.</p>

<h3>3. User Accounts</h3>
<p>When you create an account, you must provide accurate and complete information. You are responsible for maintaining the confidentiality of your account credentials.</p>

<h3>4. Purchases and Payments</h3>
<p>All purchases are subject to product availability. Prices are subject to change without notice. Payment must be received prior to shipment of products.</p>

<h3>5. Refunds and Cancellations</h3>
<p>Refund policies are determined on a case-by-case basis. Contact support for assistance with returns or cancellations.</p>

<h3>6. Intellectual Property</h3>
<p>All content, features, and functionality are owned by us and are protected by copyright, trademark, and other intellectual property laws.</p>

<h3>7. Limitation of Liability</h3>
<p>We shall not be liable for any indirect, incidental, special, consequential, or punitive damages resulting from your use of the service.</p>

<h3>8. Changes to Terms</h3>
<p>We reserve the right to modify these terms at any time. Continued use of the service constitutes acceptance of modified terms.</p>

<h3>9. Contact</h3>
<p>For questions about these Terms, please contact us.</p>'),
    ('privacy', 'Privacy Policy', '<h2>Privacy Policy</h2>
<p>Last updated: [DATE]</p>

<h3>1. Information We Collect</h3>
<p>We collect information you provide directly: name, email address, shipping address, payment information, and any other information you choose to provide.</p>

<h3>2. How We Use Your Information</h3>
<ul>
<li>Process transactions and send related information</li>
<li>Send promotional communications (with your consent)</li>
<li>Respond to your comments, questions, and requests</li>
<li>Monitor and analyze trends, usage, and activities</li>
</ul>

<h3>3. Information Sharing</h3>
<p>We do not sell, trade, or rent your personal information to third parties. We may share information with service providers who assist us in operating our platform.</p>

<h3>4. Data Security</h3>
<p>We implement appropriate security measures to protect your personal information. However, no method of transmission over the Internet is 100% secure.</p>

<h3>5. Cookies</h3>
<p>We use cookies and similar technologies to maintain your session, remember preferences, and analyze usage patterns.</p>

<h3>6. Your Rights</h3>
<p>You have the right to access, correct, or delete your personal information. Contact us to exercise these rights.</p>

<h3>7. Data Retention</h3>
<p>We retain your information for as long as your account is active or as needed to provide services and comply with legal obligations.</p>

<h3>8. Changes to This Policy</h3>
<p>We may update this Privacy Policy from time to time. We will notify you of any changes by posting the new policy on this page.</p>

<h3>9. Contact</h3>
<p>For questions about this Privacy Policy, please contact us.</p>')
ON CONFLICT (slug) DO NOTHING;
