service_account_key = out/service_account_key.json
private_key_file = .build/private_key.pem
public_key_file = .build/public_key.pem
tf_out = .build/tf.out

$(private_key_file) $(public_key_file):
	mkdir -p .build/
	openssl req -x509 -nodes -newkey rsa:2048 -days 18250 \
		-keyout "$(private_key_file)" \
		-out "$(public_key_file)" \
		-subj "/CN=unused"
	chmod 0600 "$(private_key_file)"

$(tf_out): $(public_key_file)
	mkdir -p out/
	TF_VAR_subscriber_public_key="$$(cat "$(public_key_file)")" terraform apply -input=false
	terraform output -json > "$(tf_out)"

$(service_account_key): $(private_key_file) $(tf_out)
	jq -n \
		--arg PRIVATE_KEY "$$(cat "$(private_key_file)")" \
		--arg PROJECT_ID "divera-monitor" \
		--arg CLIENT_EMAIL "$$(cat "$(tf_out)" | jq -r '.subscriber_email.value')" \
		--arg CLIENT_ID "$$(cat "$(tf_out)" | jq -r '.subscriber_id.value')" \
		--arg PRIVATE_KEY_ID "$$(cat "$(tf_out)" | jq -r '.subscriber_key_id.value')" \
		"$$(cat service_account_key.template)" \
		> "$(service_account_key)"
	chmod 0600 "$(service_account_key)"

.PHONY: build
build: $(service_account_key)
	@echo "Building..."
