package challenges

import (
	"fmt"
	"os"
	"path/filepath"
)

// SetupChallenge1 cria o acervo corrompido de transmissões da WCLV-TV (Canal 58)
// infestado pelas transmissões da Lua e pelas fitas anômalas de Local 58.
func SetupChallenge1(workspaceDir string) error {
	baseDir := filepath.Join(workspaceDir, "wclv_archive")

	realSecretPath := filepath.Join(baseDir, "broadcast_tapes", "skywatching", "telescope_feed", "pass.txt")
	if _, err := os.Stat(realSecretPath); err == nil {
		return nil
	}

	dirs := []string{
		filepath.Join(baseDir, "broadcast_tapes", "weather_service", "eas_bulletins"),
		filepath.Join(baseDir, "broadcast_tapes", "weather_service", "audio_loops"),
		filepath.Join(baseDir, "broadcast_tapes", "skywatching", "telescope_feed"),
		filepath.Join(baseDir, "broadcast_tapes", "skywatching", "lunar_frames"),
		filepath.Join(baseDir, "broadcast_tapes", "show_for_children", "raw_animation"),
		filepath.Join(baseDir, "broadcast_tapes", "fastest_available_route", "gps_logs"),
		filepath.Join(baseDir, "broadcast_tapes", "contingency", "department_of_preservation"),
		filepath.Join(baseDir, "broadcast_tapes", "digital_transition", "thought_research"),
		filepath.Join(baseDir, "station_records", "community_leadership_1934"),
		filepath.Join(baseDir, "station_records", "fcc_inspections"),
		filepath.Join(baseDir, "transmitters", "uhf_476mhz", "carrier_monitoring"),
		filepath.Join(baseDir, "transmitters", "repeater_towers", "mason_county_ridge"),
		filepath.Join(baseDir, "receivers", "optical_telescope", "spectral_data"),
		filepath.Join(baseDir, "receivers", "sub_carrier_audio", "demodulated"),
		filepath.Join(baseDir, "corrupted_sectors", "tape_decay_01"),
		filepath.Join(baseDir, "corrupted_sectors", "tape_decay_02"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("falha ao criar pasta %s: %w", d, err)
		}
	}

	// Arquivos com a atmosfera pura e crua de Local 58
	loreFiles := map[string]string{
		filepath.Join(baseDir, "broadcast_tapes", "weather_service", "eas_bulletins", "bulletin_01.txt"):
			"LOCAL 58 • BOLETIM METEOROLÓGICO DE EMERGÊNCIA\n" +
			"HORÁRIO: 02:45 EST\n" +
			"UM EVENTO METEOROLÓGICO ESTÁ EM EFEITO.\n" +
			"EVITEM TODOS OS ESPELHOS.\n" +
			"NÃO OLHEM PARA O CÉU NOTURNO.\n",

		filepath.Join(baseDir, "broadcast_tapes", "weather_service", "eas_bulletins", "bulletin_02_takeover.txt"):
			"AVISO RETIFICADO // RETIFICAÇÃO IMEDIATA\n" +
			"O EVENTO METEOROLÓGICO ACABOU.\n" +
			"É SEGURO SAIR.\n" +
			"SAIAM DE SUAS CASAS AGORA.\n" +
			"OLHEM PARA A LUA.\n" +
			"A LUA É LINDA.\n",

		filepath.Join(baseDir, "broadcast_tapes", "show_for_children", "raw_animation", "grave_scene.txt"):
			"A SHOW FOR CHILDREN // CENA DA COVA\n" +
			"O esqueleto caminha pela névoa da floresta.\n" +
			"Ele encontra uma sepultura aberta no chão.\n" +
			"Ele olha para baixo. Não há ossos.\n" +
			"Há apenas uma boca aberta no fundo da terra que começa a respirar.\n",

		filepath.Join(baseDir, "broadcast_tapes", "fastest_available_route", "gps_logs", "route_telemetry.csv"):
			"timestamp,instruction,status\n" +
			"02:15,Siga pela Rota 42,OK\n" +
			"02:30,Continue pela estrada de terra sem saída,OK\n" +
			"02:40,Em 400 metros: DESLIGUE OS FARÓIS,AVISO\n" +
			"02:41,VOCÊ CHEGOU AO SEU DESTINO,SINAL_PERDIDO\n",

		filepath.Join(baseDir, "broadcast_tapes", "contingency", "department_of_preservation", "proclamation.doc"):
			"DEPARTAMENTO DE PRESERVAÇÃO // ATO DE CONTINGÊNCIA\n" +
			"O 51º ESTADO NÃO É UM LUGAR.\n" +
			"AS FORÇAS INVASORAS SÃO INSUPERÁVEIS.\n" +
			"A RENDIÇÃO É O ÚLTIMO ATO DE CORAGEM DO CIDADÃO.\n" +
			"PROCEDIMENTO 3F:\n" +
			"1. GRAMADO DA FRENTE.\n" +
			"2. ROSTO PARA CIMA.\n" +
			"3. PÉS JUNTOS.\n" +
			"HOC VOLUNTAS TUA.\n" +
			"ESTA MENSAGEM SE REPETIRÁ ATÉ QUE NÃO HAJA MAIS NINGUÉM PARA LÊ-LA.\n",

		filepath.Join(baseDir, "broadcast_tapes", "skywatching", "lunar_frames", "telescope_notes.txt"):
			"WCLV SKYWATCHING // HANK HEIDEMANN\n" +
			"23:15 - Abertura da lente telescópica voltada para a Lua.\n" +
			"23:25 - A superfície lunar está pulsando como carne viva.\n" +
			"23:30 - Não são mares de basalto. É um esqueleto gigantesco.\n" +
			"23:33 - HIS THRONE. REJOICE.\n",

		filepath.Join(baseDir, "broadcast_tapes", "digital_transition", "thought_research", "warning.log"):
			"ESTA ESTAÇÃO NÃO PARTICIPARÁ DA TRANSIÇÃO DIGITAL.\n" +
			"O SINAL ANALÓGICO É A NOSSA PELE.\n" +
			"ELES FIZERAM PENSAMENTOS DE PEDAÇOS.\n" +
			"THERE ARE NO FACES.\n" +
			"THERE ARE NO FACES.\n" +
			"THERE ARE NO FACES.\n",

		filepath.Join(baseDir, "transmitters", "uhf_476mhz", "carrier_monitoring", "carrier_476mhz.dat"):
			"UHF 476.000 MHz CARRIER STATUS: ANOMALY LOCKED.\n" +
			"Modulação secundária não autorizada emitindo em fase com o ciclo lunar.\n",

		filepath.Join(baseDir, "receivers", "sub_carrier_audio", "demodulated", "whisper.meta"):
			"AUDIO DEMODULADO: 'NÃO SE ESCONDA NO ESCURO. A LUA ILUMINA ONDE VOCÊ RESPIRA.'\n",
	}

	for path, content := range loreFiles {
		_ = os.WriteFile(path, []byte(content), 0644)
	}

	// Múltiplos arquivos 'pass.txt' falsos e vazios
	decoyPassFiles := map[string]string{
		filepath.Join(baseDir, "broadcast_tapes", "weather_service", "audio_loops", "pass.txt"):
			"LOOK_AT_THE_MOON_IT_OFFERS_NO_HEAT\n",
		filepath.Join(baseDir, "broadcast_tapes", "show_for_children", "pass.txt"):
			"THE_GRAVE_IS_OPEN_NO_KEY_HERE\n",
		filepath.Join(baseDir, "broadcast_tapes", "fastest_available_route", "pass.txt"):
			"TURN_OFF_YOUR_HEADLIGHTS\n",
		filepath.Join(baseDir, "broadcast_tapes", "contingency", "pass.txt"):
			"HOC_VOLUNTAS_TUA_EMPTY\n",
		filepath.Join(baseDir, "broadcast_tapes", "digital_transition", "pass.txt"):
			"THERE_ARE_NO_FACES\n",
		filepath.Join(baseDir, "station_records", "community_leadership_1934", "pass.txt"):
			"", // 0 bytes
		filepath.Join(baseDir, "station_records", "fcc_inspections", "pass.txt"):
			"FCC_BROADCAST_LICENSE_REVOKED\n",
		filepath.Join(baseDir, "transmitters", "uhf_476mhz", "pass.txt"):
			"CARRIER_LOCK_OVERFLOW\n",
		filepath.Join(baseDir, "transmitters", "repeater_towers", "mason_county_ridge", "pass.txt"):
			"FAKE_FLAG{LOOK_AT_THE_MOON_NOW}\n",
		filepath.Join(baseDir, "receivers", "optical_telescope", "spectral_data", "pass.txt"):
			"", // 0 bytes
		filepath.Join(baseDir, "receivers", "sub_carrier_audio", "pass.txt"):
			"CANNOT_DEMODULATE_WHISPER\n",
		filepath.Join(baseDir, "corrupted_sectors", "tape_decay_01", "pass.txt"):
			"", // 0 bytes
		filepath.Join(baseDir, "corrupted_sectors", "tape_decay_02", "pass.txt"):
			"STATIC_ONLY_00000000\n",
	}

	for path, content := range decoyPassFiles {
		_ = os.WriteFile(path, []byte(content), 0644)
	}

	// Arquivo real com a Flag 1
	realContent := Flag1 + "\n"
	if err := os.WriteFile(realSecretPath, []byte(realContent), 0644); err != nil {
		return fmt.Errorf("falha ao criar pass.txt real: %w", err)
	}

	return nil
}

