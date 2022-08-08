package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestHearing_Markdown(t *testing.T) {
	tests := []struct {
		name    string
		hearing Hearing
		want    string
	}{
		{name: "empty", hearing: Hearing{}, want: ""},
		{
			name: "single topic",
			hearing: Hearing{
				Time:  time.Date(2022, time.March, 29, 11, 0, 0, 0, time.Local),
				Place: "г.Брянск, ул. Клинцовская, д. 60 (здание Городского Дворца культуры им. Д.Н. Медведева)",
				Topic: []string{
					"по проекту Постановления Брянской городской администрации «О предоставлении (об отказе в предоставлении) разрешений на условно разрешенный вид использования земельных участков, отклонение от предельных параметров разрешенного строительства» (далее по тексту — проект Постановления)., назначенные постановлением главы города Брянска №1151-пг от 03.03.2022 г.",
				},
				Proposals: []string{
					"Приём предложений от участников публичных слушаний, прошедших идентификацию по проекту Постановления, осуществляет оргкомитет до 28 марта 2022 года по адресу: г. Брянск, проспект Ленина, д. 28, каб. №204, в рабочие дни с 14:00 до 16:30, а 29 марта 2022 года по адресу: ул. Клинцовская, д. 60 (здание МБУК «Городской Дворец культуры им. Д.Н. Медведева») в ходе проведения публичных слушаний",
					"Приём заявлений на участие в публичных слушаниях по проекту Постановления также осуществляет оргкомитет до 28 марта 2022 года по адресу: г. Брянск, проспект Ленина, д. 28, каб. №204, в рабочие дни с 14:00 до 16:30.",
				},
				URL: "https://bga32.ru/informaciya-o-publichnyx-slushaniyax-naznachennyx-na-29-marta-2022-goda/",
			},
			want: "*29\\.03\\.2022 в 11:00 в г\\.Брянск, ул\\. Клинцовская, д\\. 60 \\(здание Городского Дворца культуры им\\. Д\\.Н\\. Медведева\\)* состоятся публичные слушания по проекту Постановления Брянской городской администрации «О предоставлении \\(об отказе в предоставлении\\) разрешений на условно разрешенный вид использования земельных участков, отклонение от предельных параметров разрешенного строительства» \\(далее по тексту — проект Постановления\\)\\., назначенные постановлением главы города Брянска №1151\\-пг от 03\\.03\\.2022 г\\.\n\nПриём предложений от участников публичных слушаний, прошедших идентификацию по проекту Постановления, осуществляет оргкомитет до 28 марта 2022 года по адресу: г\\. Брянск, проспект Ленина, д\\. 28, каб\\. №204, в рабочие дни с 14:00 до 16:30, а 29 марта 2022 года по адресу: ул\\. Клинцовская, д\\. 60 \\(здание МБУК «Городской Дворец культуры им\\. Д\\.Н\\. Медведева»\\) в ходе проведения публичных слушаний\n\nПриём заявлений на участие в публичных слушаниях по проекту Постановления также осуществляет оргкомитет до 28 марта 2022 года по адресу: г\\. Брянск, проспект Ленина, д\\. 28, каб\\. №204, в рабочие дни с 14:00 до 16:30\\.\n\n[Ссылка на публикацию](https://bga32.ru/informaciya-o-publichnyx-slushaniyax-naznachennyx-na-29-marta-2022-goda/)\n",
		},
		{
			name: "multi topics",
			hearing: Hearing{
				Time:  time.Date(2021, time.March, 17, 11, 0, 0, 0, time.Local),
				Place: "ГДК Советского района (ул. Калинина, д. 66)",
				Topic: []string{
					"по проекту планировки территории, ограниченной кольцевым пересечением в районе железнодорожного вокзала Брянск-1 территорией железнодорожного вокзала Брянск-1, руслом реки Десна и дома №19 по улице Речной в Володарском районе города Брянска",
					"по проекту внесения изменений в проект планировки и проект межевания территории, ограниченной улицами Бежицкой, Горбатова, жилой улицей № 4 в Советском районе города Брянска, в целях многоэтажного жилищного строительства в части земельных участков с кадастровыми номерами 32:28:0030902:1228, 32:28:0030902:1224, утверждённый постановлением Брянской городской администрации от 12.08.2014 №2208-п",
					"по проекту планировки, содержащему проект межевания, территории по ул. Фосфоритной, д.1 в Володарском районе города Брянска",
				},
				Proposals: []string{
					"Приём предложений от участников публичных слушаний, прошедших идентификацию по проекту Решения будет осуществлять оргкомитет до 16 марта 2021 года (включительно) по адресу: город Брянск, пр-т Ленина, д. 28, каб. №208, в рабочие дни с 14:00 до 16:30, и 17 марта 2021 года по адресу: город Брянск, улица Калинина, 66 (здание МБУК «Городской Дом культуры Советского района») в ходе проведения публичных слушаний.",
					"Приём заявлений на участие в публичных слушаниях по проекту Решения также осуществляет оргкомитет до 16 марта 2021 года (включительно) по адресу: пр-т Ленина, д. 28, каб. №208, в рабочие дни с 14.00 до 16.30.",
				},
				URL: "https://bga32.ru/informaciya-o-publichnyx-slushaniyax-naznachennyx-na-17-marta-2021-goda/",
			},
			want: "*17\\.03\\.2021 в 11:00 в ГДК Советского района \\(ул\\. Калинина, д\\. 66\\)* состоятся публичные слушания:\n\n \\- по проекту планировки территории, ограниченной кольцевым пересечением в районе железнодорожного вокзала Брянск\\-1 территорией железнодорожного вокзала Брянск\\-1, руслом реки Десна и дома №19 по улице Речной в Володарском районе города Брянска\n\n \\- по проекту внесения изменений в проект планировки и проект межевания территории, ограниченной улицами Бежицкой, Горбатова, жилой улицей № 4 в Советском районе города Брянска, в целях многоэтажного жилищного строительства в части земельных участков с кадастровыми номерами 32:28:0030902:1228, 32:28:0030902:1224, утверждённый постановлением Брянской городской администрации от 12\\.08\\.2014 №2208\\-п\n\n \\- по проекту планировки, содержащему проект межевания, территории по ул\\. Фосфоритной, д\\.1 в Володарском районе города Брянска\n\nПриём предложений от участников публичных слушаний, прошедших идентификацию по проекту Решения будет осуществлять оргкомитет до 16 марта 2021 года \\(включительно\\) по адресу: город Брянск, пр\\-т Ленина, д\\. 28, каб\\. №208, в рабочие дни с 14:00 до 16:30, и 17 марта 2021 года по адресу: город Брянск, улица Калинина, 66 \\(здание МБУК «Городской Дом культуры Советского района»\\) в ходе проведения публичных слушаний\\.\n\nПриём заявлений на участие в публичных слушаниях по проекту Решения также осуществляет оргкомитет до 16 марта 2021 года \\(включительно\\) по адресу: пр\\-т Ленина, д\\. 28, каб\\. №208, в рабочие дни с 14\\.00 до 16\\.30\\.\n\n[Ссылка на публикацию](https://bga32.ru/informaciya-o-publichnyx-slushaniyax-naznachennyx-na-17-marta-2021-goda/)\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.hearing.Markdown(), tt.name)
		})
	}
}

func TestHearing_String(t *testing.T) {
	tests := []struct {
		name    string
		hearing Hearing
		want    string
	}{
		{name: "empty", hearing: Hearing{}, want: ""},
		{
			name: "single topic",
			hearing: Hearing{

				Time:  time.Date(2022, time.March, 29, 11, 0, 0, 0, time.Local),
				Place: "г.Брянск, ул. Клинцовская, д. 60 (здание Городского Дворца культуры им. Д.Н. Медведева)",
				Topic: []string{
					"по проекту Постановления Брянской городской администрации «О предоставлении (об отказе в предоставлении) разрешений на условно разрешенный вид использования земельных участков, отклонение от предельных параметров разрешенного строительства» (далее по тексту — проект Постановления)., назначенные постановлением главы города Брянска №1151-пг от 03.03.2022 г.",
				},
				Proposals: []string{
					"Приём предложений от участников публичных слушаний, прошедших идентификацию по проекту Постановления, осуществляет оргкомитет до 28 марта 2022 года по адресу: г. Брянск, проспект Ленина, д. 28, каб. №204, в рабочие дни с 14:00 до 16:30, а 29 марта 2022 года по адресу: ул. Клинцовская, д. 60 (здание МБУК «Городской Дворец культуры им. Д.Н. Медведева») в ходе проведения публичных слушаний",
					"Приём заявлений на участие в публичных слушаниях по проекту Постановления также осуществляет оргкомитет до 28 марта 2022 года по адресу: г. Брянск, проспект Ленина, д. 28, каб. №204, в рабочие дни с 14:00 до 16:30.",
				},
				URL: "https://bga32.ru/informaciya-o-publichnyx-slushaniyax-naznachennyx-na-29-marta-2022-goda/",
			},
			want: "29.03.2022 в 11:00 в г.Брянск, ул. Клинцовская, д. 60 (здание Городского Дворца культуры им. Д.Н. Медведева) состоятся публичные слушания по проекту Постановления Брянской городской администрации «О предоставлении (об отказе в предоставлении) разрешений на условно разрешенный вид использования земельных участков, отклонение от предельных параметров разрешенного строительства» (далее по тексту — проект Постановления)., назначенные постановлением главы города Брянска №1151-пг от 03.03.2022 г.\nПриём предложений от участников публичных слушаний, прошедших идентификацию по проекту Постановления, осуществляет оргкомитет до 28 марта 2022 года по адресу: г. Брянск, проспект Ленина, д. 28, каб. №204, в рабочие дни с 14:00 до 16:30, а 29 марта 2022 года по адресу: ул. Клинцовская, д. 60 (здание МБУК «Городской Дворец культуры им. Д.Н. Медведева») в ходе проведения публичных слушаний\nПриём заявлений на участие в публичных слушаниях по проекту Постановления также осуществляет оргкомитет до 28 марта 2022 года по адресу: г. Брянск, проспект Ленина, д. 28, каб. №204, в рабочие дни с 14:00 до 16:30.\nСсылка на публикацию: https://bga32.ru/informaciya-o-publichnyx-slushaniyax-naznachennyx-na-29-marta-2022-goda/\n",
		},
		{
			name: "multi topics",
			hearing: Hearing{
				Time:  time.Date(2021, time.March, 17, 11, 0, 0, 0, time.Local),
				Place: "ГДК Советского района (ул. Калинина, д. 66)",
				Topic: []string{
					"по проекту планировки территории, ограниченной кольцевым пересечением в районе железнодорожного вокзала Брянск-1 территорией железнодорожного вокзала Брянск-1, руслом реки Десна и дома №19 по улице Речной в Володарском районе города Брянска",
					"по проекту внесения изменений в проект планировки и проект межевания территории, ограниченной улицами Бежицкой, Горбатова, жилой улицей № 4 в Советском районе города Брянска, в целях многоэтажного жилищного строительства в части земельных участков с кадастровыми номерами 32:28:0030902:1228, 32:28:0030902:1224, утверждённый постановлением Брянской городской администрации от 12.08.2014 №2208-п",
					"по проекту планировки, содержащему проект межевания, территории по ул. Фосфоритной, д.1 в Володарском районе города Брянска",
				},
				Proposals: []string{
					"Приём предложений от участников публичных слушаний, прошедших идентификацию по проекту Решения будет осуществлять оргкомитет до 16 марта 2021 года (включительно) по адресу: город Брянск, пр-т Ленина, д. 28, каб. №208, в рабочие дни с 14:00 до 16:30, и 17 марта 2021 года по адресу: город Брянск, улица Калинина, 66 (здание МБУК «Городской Дом культуры Советского района») в ходе проведения публичных слушаний.",
					"Приём заявлений на участие в публичных слушаниях по проекту Решения также осуществляет оргкомитет до 16 марта 2021 года (включительно) по адресу: пр-т Ленина, д. 28, каб. №208, в рабочие дни с 14.00 до 16.30.",
				},
				URL: "https://bga32.ru/informaciya-o-publichnyx-slushaniyax-naznachennyx-na-17-marta-2021-goda/",
			},
			want: "17.03.2021 в 11:00 в ГДК Советского района (ул. Калинина, д. 66) состоятся публичные слушания:\n - по проекту планировки территории, ограниченной кольцевым пересечением в районе железнодорожного вокзала Брянск-1 территорией железнодорожного вокзала Брянск-1, руслом реки Десна и дома №19 по улице Речной в Володарском районе города Брянска\n - по проекту внесения изменений в проект планировки и проект межевания территории, ограниченной улицами Бежицкой, Горбатова, жилой улицей № 4 в Советском районе города Брянска, в целях многоэтажного жилищного строительства в части земельных участков с кадастровыми номерами 32:28:0030902:1228, 32:28:0030902:1224, утверждённый постановлением Брянской городской администрации от 12.08.2014 №2208-п\n - по проекту планировки, содержащему проект межевания, территории по ул. Фосфоритной, д.1 в Володарском районе города Брянска\nПриём предложений от участников публичных слушаний, прошедших идентификацию по проекту Решения будет осуществлять оргкомитет до 16 марта 2021 года (включительно) по адресу: город Брянск, пр-т Ленина, д. 28, каб. №208, в рабочие дни с 14:00 до 16:30, и 17 марта 2021 года по адресу: город Брянск, улица Калинина, 66 (здание МБУК «Городской Дом культуры Советского района») в ходе проведения публичных слушаний.\nПриём заявлений на участие в публичных слушаниях по проекту Решения также осуществляет оргкомитет до 16 марта 2021 года (включительно) по адресу: пр-т Ленина, д. 28, каб. №208, в рабочие дни с 14.00 до 16.30.\nСсылка на публикацию: https://bga32.ru/informaciya-o-publichnyx-slushaniyax-naznachennyx-na-17-marta-2021-goda/\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.hearing.String(), tt.name)
		})
	}
}
